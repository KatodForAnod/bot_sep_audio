package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
)

type DownloadMod string

const (
	Common DownloadMod = "common" // download video and convert it to mp3
	// DescParts download video, convert it to mp3,
	// download description with timecodes, split mp3 by timecodes
	DescParts DownloadMod = "parts"
)

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

		command := Common
		msgText := strings.Split(update.Message.Text, " ")
		if len(msgText) > 1 {
			switch msgText[1] {
			case string(Common):
				command = Common
			case string(DescParts):
				command = DescParts
			default:
				command = Common
			}
		} else if len(msgText) == 0 {
			continue
		}

		files, err := prepareFile(msgText[0], command)
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
