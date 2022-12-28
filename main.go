package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	token := os.Getenv("telegram_token")
	if token == "" {
		log.Panic("telegram token is empty")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			file, err := prepareFile(update.Message.Text)
			if err != nil {
				log.Println(err) //send error??
				continue
			}

			_, err = bot.Send(tgbotapi.NewAudio(int64(update.Message.Chat.ID), file))
			if err != nil {
				log.Panic(err)
			}
		}
	}
}

func prepareFile(url string) (tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		log.Println(err)
		return tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = downloadMp3FileYouTube(url, tempDir)
	if err != nil {
		log.Println(err)
		return tgbotapi.FileBytes{}, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		log.Println(err)
		return tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		log.Println(err)
		return tgbotapi.FileBytes{}, errors.New("no files")
	}

	fileName := filepath.Join(tempDir, fileInfo[0].Name())
	mp3File, err := os.ReadFile(fileName)
	if err != nil {
		log.Println(err)
		return tgbotapi.FileBytes{}, err
	}

	mp3FileBytes := tgbotapi.FileBytes{
		Name:  fileInfo[0].Name(),
		Bytes: mp3File,
	}

	return mp3FileBytes, nil
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
