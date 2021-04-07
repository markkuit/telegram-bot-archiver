package main

import (
	"encoding/json"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/markkuit/telegram-bot-archiver/internal/callback"
	"github.com/markkuit/telegram-bot-archiver/internal/command"
	"github.com/markkuit/telegram-bot-archiver/internal/i18n"
	"github.com/markkuit/telegram-bot-archiver/internal/session"
	"github.com/markkuit/telegram-bot-archiver/internal/telegram"
)

func main() {
	for in := range telegram.UpdatesChannel {
		if in.CallbackQuery != nil {
			var callbackData callback.Callback
			if err := json.Unmarshal([]byte(in.CallbackQuery.Data), &callbackData); err != nil {
				log.Fatalf("callbackData: Unmarshal: %s\n", err)
			}

			switch callbackData.Command {
			case callback.CommandExtractChoose:
				go command.ExtractChoose(in.CallbackQuery, callbackData)
			case callback.CommandExtractDone:
				go command.ExtractDone(in.CallbackQuery)
			}

			telegram.AcknowledgeCallbackQuery(in.CallbackQuery)
			continue
		}

		if in.Message != nil {
			if in.Message.IsCommand() {
				switch in.Message.Command() {
				case "start":
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.Usage, ParseMode: tgbotapi.ModeMarkdown})
				case "cancel":
					session, exists := session.Session(in.Message.Chat.ID)
					if !exists {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorNoSession})
						continue
					}

					session.Delete()
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.SessionDeleted})
				case "compress":
					if session.Exists(in.Message.Chat.ID) {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.SessionAlreadyActive, ParseMode: tgbotapi.ModeMarkdown})
						continue
					}
					if in.Message.CommandArguments() == "" {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorArgumentMissing, ParseMode: tgbotapi.ModeMarkdown})
						continue
					}
					go command.Compress(in.Message)
				case "compressdone":
					go command.CompressDone(in.Message)
				case "extractfile":
					if session.Exists(in.Message.Chat.ID) {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.SessionAlreadyActive, ParseMode: tgbotapi.ModeMarkdown})
						continue
					}

					session.New(in.Message.Chat.ID, "extractfile", "")
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ExtractSendArchive})
				case "extracturl":
					if session.Exists(in.Message.Chat.ID) {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.SessionAlreadyActive, ParseMode: tgbotapi.ModeMarkdown})
						continue
					}
					if in.Message.CommandArguments() == "" {
						telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorArgumentMissing, ParseMode: tgbotapi.ModeMarkdown})
						continue
					}
					go command.ExtractURL(in.Message)
				default:
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorUnrecognizedCommand, ParseMode: tgbotapi.ModeMarkdown})
				}
				continue
			}

			if in.Message.Document != nil {
				session, exists := session.Session(in.Message.Chat.ID)
				if !exists {
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorNoSession})
					continue
				}
				if !(session.Action == "extractfile" || session.Action == "compress") {
					telegram.SendMessage(telegram.Message{ChatID: in.Message.Chat.ID, Text: i18n.ErrorUnexpectedDocument, ParseMode: tgbotapi.ModeMarkdown})
					continue
				}

				switch session.Action {
				case "compress":
					go command.CompressAddItem(in.Message)
				case "extractfile":
					go command.ExtractFile(in.Message)
				}
				continue
			}
		}
	}
}
