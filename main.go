package main

import (
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
	//LogInit()
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

		files, err := prepareFile(update.Message.Text)
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

func splitLongMP3(longMP3PathFile string) ([]tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = splitLongAudioCmd(tempDir, longMP3PathFile)
	if err != nil {
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

func prepareFile(url string) ([]tgbotapi.FileBytes, error) {
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
			"main.go", "prepareFile", "ioutil.ReadDir", errors.New("no files"))
		return []tgbotapi.FileBytes{}, err
	}

	maxSize := 1024 * 1024 * 50
	if fileInfo[0].Size() < int64(maxSize) {
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
	}

	mp3ShortFiles, err := splitLongMP3(filepath.Join(tempDir, fileInfo[0].Name()))
	if err != nil {
		return []tgbotapi.FileBytes{}, err
	}

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

	return nil
}

func splitLongAudioCmd(dirForSplitAudio, longAudioFilePath string) error {
	name := path.Base(longAudioFilePath)
	cmd := exec.Command("ffmpeg", "-i", longAudioFilePath,
		"-acodec", "copy", "-vn", "-f", "segment", "-segment_time", "2700",
		filepath.Join(dirForSplitAudio, "Part %d."+name))
	err := cmd.Run()
	if err != nil {
		/*err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
		"main.go", "downloadMp3FileYouTube", "exec.Command", err)*/
		log.Println(err)
		return err
	}

	return nil
}
