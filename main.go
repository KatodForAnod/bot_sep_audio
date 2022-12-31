package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	//u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		file, err := prepareFile(update.Message.Text)
		if err != nil {
			log.Println(err)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"something went wrong!"))
			continue
		}
		_, err = bot.Send(tgbotapi.NewAudio(update.Message.Chat.ID, file))
		if err != nil &&
			err.Error() == http.StatusText(http.StatusRequestEntityTooLarge) {
			smallerFiles, err := splitLongMP3(file)
			if err != nil {
				log.Println(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"something went wrong!"))
				continue
			}
			for _, smallerFile := range smallerFiles {
				bot.Send(tgbotapi.NewAudio(update.Message.Chat.ID, smallerFile))
			}
		} else {
			log.Println(err)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"something went wrong!"))
		}
	}
}

func splitLongMP3(longMP3 tgbotapi.FileBytes) ([]tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		log.Println(err)
		return []tgbotapi.FileBytes{}, err
	}

	tempFile, err := os.CreateTemp(tempDir, "longAudio.*.mp3")
	if err != nil {
		return []tgbotapi.FileBytes{}, err
	}
	if _, err := tempFile.Write(longMP3.Bytes); err != nil {
		log.Fatal(err)
	}
	if err := tempFile.Close(); err != nil {
		log.Fatal(err)
	}

	err = splitLongAudio(tempDir, tempFile.Name())
	if err != nil {
		return nil, err
	}

	os.Remove(tempFile.Name())

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

func prepareFile(url string) (tgbotapi.FileBytes, error) {
	tempDir, err := os.MkdirTemp("", "tempDir")
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "os.MkdirTemp", err)
		return tgbotapi.FileBytes{}, err
	}
	defer os.RemoveAll(tempDir)

	err = downloadMp3FileYouTube(url, tempDir)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "downloadMp3FileYouTube", err)
		return tgbotapi.FileBytes{}, err
	}

	fileInfo, err := ioutil.ReadDir(tempDir)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "ioutil.ReadDir", err)
		return tgbotapi.FileBytes{}, err
	}
	if len(fileInfo) == 0 {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "ioutil.ReadDir", errors.New("no files"))
		return tgbotapi.FileBytes{}, err
	}

	fileName := filepath.Join(tempDir, fileInfo[0].Name())
	mp3File, err := os.ReadFile(fileName)
	if err != nil {
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "prepareFile", "os.ReadFile", err)
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
		err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
			"main.go", "downloadMp3FileYouTube", "exec.Command", err)
		return err
	}

	return nil
}

func splitLongAudio(dir, fileName string) error {
	// ffmpeg -i long.mp3 -acodec copy -vn -f segment -segment_time 30 half%d.mp3
	fmt.Println("test:", fileName)
	fmt.Println("test2:", filepath.Join(dir, "%d.mp3"))
	cmd := exec.Command("ffmpeg", "-i", fileName,
		"-acodec", "copy", "-vn", "-f", "segment", "-segment_time", "2700", filepath.Join(dir, "%d.mp3"))
	err := cmd.Run()
	if err != nil {
		/*err = fmt.Errorf("file: %s, func: %s, action: %s, error: %w",
		"main.go", "downloadMp3FileYouTube", "exec.Command", err)*/
		log.Println(err)
		return err
	}

	return nil
}
