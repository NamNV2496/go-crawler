package service

import (
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
	ChatId      int64
	ChannelName string
	bot         *tgbotapi.BotAPI
}

func NewTeleService(
	conf *configs.Config,
) *TeleService {
	bot, err := tgbotapi.NewBotAPI(conf.Telegram.APIKey)
	if err != nil {
		panic(err)
	}

	return &TeleService{
		ChatId:      conf.Telegram.ChatId,
		ChannelName: conf.Telegram.ChannelName,
		bot:         bot,
	}
}

var _ ITeleService = &TeleService{}

func (s *TeleService) SendNotify(telephone string) bool {
	msg := tgbotapi.NewMessage(s.bot.Self.ID, telephone)
	_, err := s.bot.Send(msg)
	if err != nil {
		return false
	}
	return true
}

func (s *TeleService) SendMessage(message string, format string) error {
	if len(message) == 0 {
		fmt.Println("Message is empty => Not sending message")
		return nil
	}
	var msg tgbotapi.MessageConfig
	if s.ChatId != 0 {
		msg = tgbotapi.NewMessage(s.ChatId, message)
	} else if len(s.ChannelName) != 0 {
		msg = tgbotapi.NewMessageToChannel(s.ChannelName, message)
	} else {
		os.Exit(1)
	}
	if format == "markdown" {
		msg.ParseMode = tgbotapi.ModeMarkdown
	} else if format == "html" {
		msg.ParseMode = tgbotapi.ModeHTML
	}
	_, err := s.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *TeleService) SendLocation(latitude float64, longitude float64) error {
	if longitude < -180 || longitude > 180 || latitude < -90 || latitude > 90 {
		fmt.Printf("Longitude or latitude value invalid: %v, %v\n", latitude, longitude)
		os.Exit(1)
	}

	msg := tgbotapi.NewLocation(s.ChatId, latitude, longitude)
	_, err := s.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (s *TeleService) SendFile(filename string, filetype string, caption string) error {
	file := fileReader(filename, filetype, caption)
	var msg tgbotapi.Chattable
	switch filetype {
	case "photo":
		msg = tgbotapi.NewPhoto(s.ChatId, file)
	case "video":
		msg = tgbotapi.NewVideo(s.ChatId, file)
	case "audio":
		msg = tgbotapi.NewAudio(s.ChatId, file)
	case "sticker":
		msg = tgbotapi.NewSticker(s.ChatId, file)
	case "animation":
		msg = tgbotapi.NewAnimation(s.ChatId, file)
	default:
		msg = tgbotapi.NewDocument(s.ChatId, file)
	}

	_, err := s.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func fileReader(filename string, filetype string, caption string) (file tgbotapi.FileReader) {
	reader, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Reading file %v error: %v\n", filename, err.Error())
		os.Exit(1)
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
		os.Exit(1)
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
		os.Exit(1)
	}
}
