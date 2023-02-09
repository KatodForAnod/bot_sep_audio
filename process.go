package main

import (
	"bot_sep_audio/parser"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func preparePartsMP3(longMP3PathFile string,
	splitFunc func(dirForSplitAudio string, mainAudio string) error) ([]tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = splitFunc(tempDir, longMP3PathFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}

	var splitFiles []tgbotapi.FileBytes
	for _, info := range fileInfo {
		fileName := filepath.Join(tempDir, info.Name())
		mp3File, err := os.ReadFile(fileName)
		if err != nil {
			log.Println(err)
			return []tgbotapi.FileBytes{}, err
		}

		mp3FileBytes := tgbotapi.FileBytes{
			Name:  info.Name(),
			Bytes: mp3File,
		}
		splitFiles = append(splitFiles, mp3FileBytes)
	}

	return splitFiles, nil
}

func prepareFile(url string, mod DownloadMod) ([]tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = downloadMp3FileYouTube(url, tempDir)
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}

	var mp3ShortFiles []tgbotapi.FileBytes
	switch mod {
	case Common:
		maxSize := 1024 * 1024 * 50
		if fileInfo[0].Size() < int64(maxSize) {
			fileName := filepath.Join(tempDir, fileInfo[0].Name())
			mp3File, err := os.ReadFile(fileName)
			if err != nil {
				log.Println(err)
				return []tgbotapi.FileBytes{}, err
			}
			return []tgbotapi.FileBytes{
				{Name: fileInfo[0].Name(), Bytes: mp3File},
			}, nil
		}
		mp3ShortFiles, err = preparePartsMP3(filepath.Join(tempDir, fileInfo[0].Name()), splitLongAudioCmd)
		if err != nil {
			return []tgbotapi.FileBytes{}, err
		}
	case DescParts:
		splitConfigParts := func(dirForSplitAudio, longAudioFilePath string) error {
			info, err := parser.GetVideoPartsInfo(url)
			if err != nil {
				log.Println(err)
				return err
			}
			err = splitAudioToPartsCmd(dirForSplitAudio, longAudioFilePath, info)
			if err != nil {
				log.Println(err)
				return err
			}
			return nil
		}
		mp3ShortFiles, err = preparePartsMP3(filepath.Join(tempDir, fileInfo[0].Name()), splitConfigParts)
		if err != nil {
			return []tgbotapi.FileBytes{}, err
		}
	}

	return mp3ShortFiles, nil
}

func downloadMp3FileYouTube(url, dir string) error {
	cmd := exec.Command("yt-dlp", "-f", "22",
		"-x", "--audio-format", "mp3", "-o", filepath.Join(dir, "%(title)s.%(ext)s"), url)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func splitLongAudioCmd(dirForSplitAudio, longAudioFilePath string) error {
	name := path.Base(longAudioFilePath)
	cmd := exec.Command("ffmpeg", "-i", longAudioFilePath,
		"-acodec", "copy", "-vn", "-f", "segment", "-segment_time", "2700",
		filepath.Join(dirForSplitAudio, "Part %d."+name))
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func splitAudioToPartsCmd(dirForPartsAudio, audioFilePath string, arr []parser.VideoParts) error {
	for i := 0; i < len(arr)-1; i++ {
		cmd := exec.Command("ffmpeg", "-i", audioFilePath,
			"-ss", arr[i].Start, "-to", arr[i+1].Start, "-c", "copy",
			filepath.Join(dirForPartsAudio, arr[i].Name+".mp3"))
		err := cmd.Run()
		if err != nil {
			log.Println(err)
			continue
		}
	}

	cmd := exec.Command("ffmpeg", "-i", audioFilePath,
		"-ss", arr[len(arr)-1].Start, "-c", "copy",
		filepath.Join(dirForPartsAudio, arr[len(arr)-1].Name+".mp3"))
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	//return error
	return nil
}
