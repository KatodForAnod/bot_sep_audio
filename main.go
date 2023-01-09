package main

import (
	"bot_sep_audio/parser"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DownloadMod string

const (
	Common DownloadMod = "common" // download video and convert it to mp3
	// DescParts download video, convert it to mp3,
	// download description with timecodes, split mp3 by timecodes
	DescParts DownloadMod = "parts"
)

const dirName = "logs"

func LogInit() error {
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fileLogName := time.Now().Format("2006-01-02") + ".txt"
	f, err := os.OpenFile(dirName+"/"+fileLogName, os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	token := os.Getenv("telegram_token")
	if token == "" {
		log.Panic("telegram token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		files, err := prepareFile(update.Message.Text, Common)
		if err != nil {
			log.Println(err)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"something went wrong!"))
			continue
		}

		for _, filesMp3 := range files {
			_, err = bot.Send(tgbotapi.NewAudio(update.Message.Chat.ID, filesMp3))
			if err != nil {
				log.Println(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"something went wrong!"))
			}
		}
	}
}

func preparePartsMP3(longMP3PathFile string,
	splitFunc func(dirForSplitAudio string, mainAudio string) error) ([]tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "preparePartsMP3", "os.MkdirTemp", err)
		return []tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = splitFunc(tempDir, longMP3PathFile)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "preparePartsMP3", "splitLongAudioCmd", err)
		return nil, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "preparePartsMP3", "ioutil.ReadDir", err)
		return []tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "preparePartsMP3", "len(fileInfo) == 0", errors.New("no files"))
		return []tgbotapi.FileBytes{}, err
	}

	var splitFiles []tgbotapi.FileBytes
	for _, info := range fileInfo {
		fileName := filepath.Join(tempDir, info.Name())
		mp3File, err := os.ReadFile(fileName)
		if err != nil {
			err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
				"main.go", "preparePartsMP3", "os.ReadFile", err)
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
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "os.MkdirTemp", err)
		return []tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = downloadMp3FileYouTube(url, tempDir)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "downloadMp3FileYouTube", err)
		return []tgbotapi.FileBytes{}, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "ioutil.ReadDir", err)
		return []tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "len info", errors.New("no files"))
		return []tgbotapi.FileBytes{}, err
	}

	/*maxSize := 1024 * 1024 * 50
	if fileInfo[0].Size() < int64(maxSize) && mod == common {
		fileName := filepath.Join(tempDir, fileInfo[0].Name())
		mp3File, err := os.ReadFile(fileName)
		if err != nil {
			err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
				"main.go", "prepareFile", "os.ReadFile", err)
			return []tgbotapi.FileBytes{}, err
		}
		return []tgbotapi.FileBytes{
			{Name: fileInfo[0].Name(), Bytes: mp3File},
		}, nil
	}*/

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
	mp3ShortFiles, err := preparePartsMP3(filepath.Join(tempDir, fileInfo[0].Name()), splitConfigParts)
	if err != nil {
		return []tgbotapi.FileBytes{}, err
	}

	/*mp3ShortFiles, err = preparePartsMP3(filepath.Join(tempDir, fileInfo[0].Name()), splitLongAudioCmd)
	if err != nil {
		return []tgbotapi.FileBytes{}, err
	}*/

	return mp3ShortFiles, nil
}

func downloadMp3FileYouTube(url, dir string) error {
	cmd := exec.Command("yt-dlp", "-f", "22",
		"-x", "--audio-format", "mp3", "-o", filepath.Join(dir, "%(title)s.%(ext)s"), url)
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "downloadMp3FileYouTube", "exec.Command", err)
		return err
	}
	fmt.Println(cmd.String())
	return nil
}

func splitLongAudioCmd(dirForSplitAudio, longAudioFilePath string) error {
	name := path.Base(longAudioFilePath)
	cmd := exec.Command("ffmpeg", "-i", longAudioFilePath,
		"-acodec", "copy", "-vn", "-f", "segment", "-segment_time", "2700",
		filepath.Join(dirForSplitAudio, "Part %d."+name))
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "splitLongAudioCmd", "exec.Command", err)
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
