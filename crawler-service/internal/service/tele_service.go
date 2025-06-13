package service

import (
	"errors"
	"fmt"
	"os"
	"path"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/namnv2496/crawler/internal/configs"
)

type ITeleService interface {
	SendMessage(message string, format string) error
	SendLocation(latitude float64, longitude float64) error
	SendFile(filename string, filetype string, caption string) error
}

type TeleService struct {
	enable      bool
	chatId      int64
	channelName string
	bot         *tgbotapi.BotAPI
}

func NewTeleService(
	conf *configs.Config,
) *TeleService {
	if !conf.Telegram.Enable {
		return &TeleService{}
	}
	bot, err := tgbotapi.NewBotAPI(conf.Telegram.APIKey)
	if err != nil {
		panic(err)
	}

	return &TeleService{
		enable:      conf.Telegram.Enable,
		chatId:      conf.Telegram.ChatId,
		channelName: conf.Telegram.ChannelName,
		bot:         bot,
	}
}

var _ ITeleService = &TeleService{}

func (_self *TeleService) SendNotify(telephone string) bool {
	msg := tgbotapi.NewMessage(_self.bot.Self.ID, telephone)
	_, err := _self.bot.Send(msg)
	if err != nil {
		return false
	}
	return true
}

func (_self *TeleService) SendMessage(message string, format string) error {
	if !_self.enable {
		return nil
	}
	if len(message) == 0 {
		fmt.Println("Message is empty => Not sending message")
		return nil
	}
	var msg tgbotapi.MessageConfig
	if _self.chatId != 0 {
		msg = tgbotapi.NewMessage(_self.chatId, message)
	} else if len(_self.channelName) != 0 {
		msg = tgbotapi.NewMessageToChannel(_self.channelName, message)
	} else {
		return errors.New("chatId and channelName are empty")
	}
	if format == "markdown" {
		msg.ParseMode = tgbotapi.ModeMarkdown
	} else if format == "html" {
		msg.ParseMode = tgbotapi.ModeHTML
	}
	_, err := _self.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (_self *TeleService) SendLocation(latitude float64, longitude float64) error {
	if !_self.enable {
		return nil
	}
	if longitude < -180 || longitude > 180 || latitude < -90 || latitude > 90 {
		fmt.Printf("Longitude or latitude value invalid: %v, %v\n", latitude, longitude)
		return errors.New("invalid longitude or latitude value")
	}

	msg := tgbotapi.NewLocation(_self.chatId, latitude, longitude)
	_, err := _self.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (_self *TeleService) SendFile(filename string, filetype string, caption string) error {
	if !_self.enable {
		return nil
	}
	file := fileReader(filename, filetype, caption)
	var msg tgbotapi.Chattable
	switch filetype {
	case "photo":
		msg = tgbotapi.NewPhoto(_self.chatId, file)
	case "video":
		msg = tgbotapi.NewVideo(_self.chatId, file)
	case "audio":
		msg = tgbotapi.NewAudio(_self.chatId, file)
	case "sticker":
		msg = tgbotapi.NewSticker(_self.chatId, file)
	case "animation":
		msg = tgbotapi.NewAnimation(_self.chatId, file)
	default:
		msg = tgbotapi.NewDocument(_self.chatId, file)
	}

	_, err := _self.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func fileReader(filename string, filetype string, caption string) (file tgbotapi.FileReader) {
	reader, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Reading file %v error: %v\n", filename, err.Error())
		return
	}

	stat, _ := reader.Stat()
	size(filename, stat, filetype)

	if caption == "" {
		caption = path.Base(filename)
	}

	file = tgbotapi.FileReader{
		Name:   caption,
		Reader: reader,
	}

	return file
}

func size(filename string, fileInfo os.FileInfo, filetype string) {
	if fileInfo.IsDir() {
		fmt.Printf("Error: '%v' is a directory.\n", filename)
		return
	}

	var sizeLimit int64
	switch filetype {
	case "photo":
		sizeLimit = 10 * 1024 * 1024 // image max size is 10M.
	default:
		sizeLimit = 50 * 1024 * 1024 // Telegram bot api limit file size to 50MB.
	}

	fileSize := fileInfo.Size()
	if fileSize > sizeLimit {
		fmt.Printf("File %v is too large, size: %.2f MB, size limit: %v MB\n",
			filename, float64(fileSize)/(1024*1024), sizeLimit/(1024*1024))
		return
	}
}
