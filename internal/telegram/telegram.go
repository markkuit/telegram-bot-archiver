package telegram

import (
	"errors"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/markkuit/telegram-bot-archiver/internal/commons"
	"github.com/markkuit/telegram-bot-archiver/internal/i18n"
)

var bot *tgbotapi.BotAPI
var UpdatesChannel tgbotapi.UpdatesChannel

type Message struct {
	ChatID         int64
	Text           string
	ParseMode      string
	KeyboardMarkup interface{}
}

func init() {
	var err error
	bot, err = tgbotapi.NewBotAPI(commons.Config.APIToken)
	if err != nil {
		log.Fatalf("init: NewBotAPI: %s\n", err)
	}
	log.Printf("Login successful: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	UpdatesChannel, err = bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("init: GetUpdatesChan: %s\n", err)
	}
}

// AcknowledgeCallbackQuery sends back a (empty) callback ACK for the given query
func AcknowledgeCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) error {
	if _, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, "")); err != nil {
		return err
	}
	return nil
}

// ClearKeyboardMarkup assigns an empty keyboard markup to the given message
func ClearKeyboardMarkup(message *tgbotapi.Message) error {
	if _, err := bot.Send(
		tgbotapi.NewEditMessageReplyMarkup(
			message.Chat.ID, message.MessageID, tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: make([][]tgbotapi.InlineKeyboardButton, 0),
			},
		)); err != nil {
		return err
	}
	return nil
}

// GetFileDirectURL wraps Bot API direct file download for size checking
// Note that tgbotapi fails if the maximum protocol size is exceeded
func GetFileDirectURL(message *tgbotapi.Message) (string, error) {
	if message.Document.FileSize > int(commons.Config.MaxSingleFileSize) {
		return "", fmt.Errorf("GetFileDirectURL: %s: %s", i18n.ErrorMaxSingleFileSizeExceeded, message.Document.FileName)
	}

	directURL, err := bot.GetFileDirectURL(message.Document.FileID)
	if err != nil {
		return "", err
	}
	return directURL, nil
}

// SendMessage wraps Bot API message sending for less boilerplate
func SendMessage(message Message) error {
	if message.ChatID == 0 || message.Text == "" {
		return errors.New("SendMessage: ChatID and Text are mandatory")
	}

	msg := tgbotapi.NewMessage(message.ChatID, message.Text)
	msg.ParseMode = message.ParseMode
	msg.ReplyMarkup = message.KeyboardMarkup
	if _, err := bot.Send(msg); err != nil {
		log.Printf("SendMessage: Send: %s\n", err)
	}
	return nil
}

// SendFile wraps Bot API file sending for less boilerplate
func SendFile(chatID int64, file interface{}) error {
	var msg tgbotapi.DocumentConfig
	switch s := file.(type) {
	case tgbotapi.FileReader, string:
		msg = tgbotapi.NewDocumentUpload(chatID, file)
	default:
		return fmt.Errorf("SendFile: unsupported type: %T", s)
	}

	if msg.FileSize > int(commons.Config.MaxSingleFileSize) {
		return errors.New("SendFile: " + i18n.ErrorMaxSingleFileSizeExceeded)
	}

	if _, err := bot.Send(msg); err != nil {
		log.Printf("SendFile: Send: %s\n", err)
	}
	return nil
}
