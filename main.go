package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
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

		audioConfigs, err := processChatUpdate(update)
		if err != nil {
			log.Println(err)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
				"something went wrong!"))
		}
		for _, file := range audioConfigs {
			_, err = bot.Send(tgbotapi.NewAudio(update.Message.Chat.ID, file))
			if err != nil {
				log.Println(err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
					"something went wrong!"))
			}
		}
	}
}
