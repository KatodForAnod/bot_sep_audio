package lib

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"strings"
)

type TelegramBot struct {
}

const (
	startBotCommand = "/start"
)

func (t *TelegramBot) telegramCommands(botCommand string) string {
	switch botCommand {
	case startBotCommand:
		return "Данный бот предназначен для скачивания аудио из ютуба." +
			" Для этого нужно отправить в сообщении ссылку на интересующее вас видео," +
			" бот его обработает и извлечет аудио." +
			" Бот находятся в бета версии, круг пользователей на данный момент ОГРАНИЧЕН." +
			" Обратная связь: telegram@katodForAnod"
	default:
		return "Неизвестная команда"
	}
}

func (t *TelegramBot) ParsePrepareMessage(update tgbotapi.Update) ([]tgbotapi.Chattable, error) {
	msgText := strings.Split(update.Message.Text, " ")

	if len(msgText) == 0 {
		log.Printf("user: %s, error code: %d, str: %s",
			update.Message.Chat.UserName, SomethingWentWrong, BotErrorToStr[SomethingWentWrong])
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(update.Message.Chat.ID, BotErrorToStr[SomethingWentWrong])}, nil
	}

	if len(msgText) == 1 && len(msgText[0]) != 0 {
		if msgText[0][0:1] == "/" {
			answer := t.telegramCommands(msgText[0])
			return []tgbotapi.Chattable{
				tgbotapi.NewMessage(update.Message.Chat.ID, answer)}, nil
		}
	}

	if !t.isUserHaveAccess(update.Message.Chat.UserName) {
		log.Printf("user: %s dont have access, error code: %d",
			update.Message.Chat.UserName, UserDontHaveAccess)
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(update.Message.Chat.ID, BotErrorToStr[UserDontHaveAccess])}, nil
	}

	_, err := url.ParseRequestURI(msgText[0])
	if err != nil {
		log.Printf("user: %s, wrong url: %s, err %s",
			update.Message.Chat.UserName, msgText[0], err)
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(update.Message.Chat.ID, BotErrorToStr[UnCorrectUrl])}, nil
	}

	audioConfigs, err := ProcessChatUpdate(update)
	if err != nil {
		log.Printf("user: %s, ProcessChatUpdate error: %s",
			update.Message.Chat.UserName, err)
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(update.Message.Chat.ID, BotErrorToStr[SomethingWentWrong])}, nil
	}

	var preparedAudio []tgbotapi.Chattable
	for _, file := range audioConfigs {
		preparedAudio = append(preparedAudio, tgbotapi.NewAudio(update.Message.Chat.ID, file))
	}
	return preparedAudio, nil
}

func (t *TelegramBot) isUserHaveAccess(username string) bool {
	return false
}
