package command

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/markkuit/telegram-bot-archiver/internal/commons"
	"github.com/markkuit/telegram-bot-archiver/internal/i18n"
	"github.com/markkuit/telegram-bot-archiver/internal/session"
	"github.com/markkuit/telegram-bot-archiver/internal/telegram"
	"github.com/markkuit/telegram-bot-archiver/internal/util"
	"github.com/mholt/archiver"
)

// Compress handles the initialization of a compress session, making sure
// the format is supported by the library in use
func Compress(message *tgbotapi.Message) {
	filePath := filepath.Join(commons.Config.TempPath, strconv.Itoa(int(message.Chat.ID))+"."+message.CommandArguments())
	if _, err := archiver.ByExtension(filePath); err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.CompressInitializationFailed})
		return
	}

	session.New(message.Chat.ID, "compress", filePath)
	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.CompressSendFiles, ParseMode: tgbotapi.ModeMarkdown})
}

// CompressAddItem handles the addition of files to the current session archive, upon
// download of the file itself from a Telegram document
func CompressAddItem(message *tgbotapi.Message) {
	session, exists := session.Session(message.Chat.ID)
	if !exists {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.ErrorNoSession})
		return
	}

	directURL, err := telegram.GetFileDirectURL(message)
	if err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorGetFileFailed, err.Error())})
		return
	}

	filePath := filepath.Join(commons.Config.TempPath, strconv.Itoa(int(message.Chat.ID)), message.Document.FileName)
	if err := util.DownloadFile(filePath, directURL); err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorDownloadFileFailed, err.Error())})
		return
	}

	session.AddSessionFile(filePath)
	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.CompressProcessedFile, message.Document.FileName)})
}

// CompressDone ends the current compress session, if any
func CompressDone(message *tgbotapi.Message) {
	session, exists := session.Session(message.Chat.ID)
	if !exists {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.ErrorNoSession})
		return
	}
	defer session.Delete()

	archiverInterfaces, err := archiver.ByExtension(session.ActionFile)
	if err != nil {
		log.Printf("CompressDone: ByExtension: %s\n", err)
		return
	}
	archiver := archiverInterfaces.(archiver.Archiver)
	if err := archiver.Archive(session.SessionFiles, session.ActionFile); err != nil {
		log.Printf("CompressDone: Archive: %s\n", err)
		return
	}

	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.SendingFile})
	if err := telegram.SendFile(message.Chat.ID, session.ActionFile); err != nil {
		log.Printf("CompressDone: SendFile: %s\n", err)
		return
	}
	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.SessionOver})
}
