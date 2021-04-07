package command

import (
	"archive/tar"
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/dustin/go-humanize"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/markkuit/telegram-bot-archiver/internal/callback"
	"github.com/markkuit/telegram-bot-archiver/internal/commons"
	"github.com/markkuit/telegram-bot-archiver/internal/i18n"
	"github.com/markkuit/telegram-bot-archiver/internal/session"
	"github.com/markkuit/telegram-bot-archiver/internal/telegram"
	"github.com/markkuit/telegram-bot-archiver/internal/util"
	"github.com/mholt/archiver"
	"github.com/nwaples/rardecode"
)

// buildCallbackRows builds slices of buttons based on a list of archive files
func buildCallbackRows(archiveFiles map[int]archiver.File) ([][]tgbotapi.InlineKeyboardButton, error) {
	var callbackButtons []tgbotapi.InlineKeyboardButton
	for k, v := range archiveFiles {
		if v.Mode().IsRegular() && v.Size() > 0 {
			callback := callback.Callback{
				Command: callback.CommandExtractChoose,
				Index:   []int{k},
			}
			callbackMarshal, err := json.Marshal(callback)
			if err != nil {
				return [][]tgbotapi.InlineKeyboardButton{}, err
			}

			// not all formats support getting the full path, thus having to check by the header
			var fileName string
			switch h := v.Header.(type) {
			case zip.FileHeader:
				fileName = h.Name
			case *tar.Header:
				fileName = h.Name
			case *rardecode.FileHeader:
				fileName = h.Name
			default:
				fileName = v.Name()
			}

			callbackButtons = append(callbackButtons, tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s (%s)", fileName, humanize.Bytes(uint64(v.Size()))),
				string(callbackMarshal),
			))
		}
	}

	// ref. https://stackoverflow.com/a/35179941
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(callbackButtons); i += commons.Config.InlineButtonsRowSize {
		end := i + commons.Config.InlineButtonsRowSize

		if end > len(callbackButtons) {
			end = len(callbackButtons)
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(callbackButtons[i:end]...))
	}

	return rows, nil
}

// buildExtract builds the actual session buttons to let the user choose
// which files to download, or all of them, assuming there are valid ones
func buildExtract(filePath, sessionAction string, message *tgbotapi.Message) {
	archiveFiles, err := util.ArchiveWalk(filePath)
	if err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorArchiveWalkFailed, err.Error())})
		return
	}
	inlineRows, err := buildCallbackRows(archiveFiles)
	if err != nil {
		log.Fatalf("buildExtract: buildCallbackRows: %s\n", err)
	}

	if len(inlineRows) > 0 {
		var callbackData callback.Callback

		callbackData = callback.Callback{
			Command: callback.CommandExtractChoose,
			Flags:   []int{callback.FlagWildcard},
		}
		callbackMarshalGetAll, err := json.Marshal(callbackData)
		if err != nil {
			log.Printf("SendMessage: Send: %s\n", err)
		}

		callbackData = callback.Callback{
			Command: callback.CommandExtractDone,
		}
		callbackMarshalDone, err := json.Marshal(callbackData)
		if err != nil {
			log.Fatalf("callbackData: Marshal: %s\n", err)
		}

		inlineRows = append(inlineRows,
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Get all", string(callbackMarshalGetAll)),
				tgbotapi.NewInlineKeyboardButtonData("Done", string(callbackMarshalDone)),
			),
		)
		inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(inlineRows...)

		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.ExtractChooseFiles, KeyboardMarkup: inlineKeyboard})
		session.New(message.Chat.ID, sessionAction, filePath)
	} else {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.ExtractNoValidFiles})
	}
}

// ExtractFile handles the extraction of specific archive files or all of them, upon
// download of the archive itself from a Telegram document
func ExtractFile(message *tgbotapi.Message) {
	directURL, err := telegram.GetFileDirectURL(message)
	if err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorGetFileFailed, err.Error())})
		return
	}

	filePath := filepath.Join(commons.Config.TempPath, strconv.Itoa(int(message.Chat.ID))+message.Document.FileName)
	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.DownloadingFile})
	if err := util.DownloadFile(filePath, directURL); err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorDownloadFileFailed, err.Error())})
		return
	}

	buildExtract(filePath, "extractfile", message)
}

// ExtractFile handles the extraction of specific archive files or all of them, upon
// download of the archive itself from a common URL
func ExtractURL(message *tgbotapi.Message) {
	srcURL, err := url.Parse(message.CommandArguments())
	if err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorURLParseFailed, err.Error())})
		return
	}

	filePath := filepath.Join(commons.Config.TempPath, strconv.Itoa(int(message.Chat.ID))+filepath.Base(srcURL.Path))
	telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: i18n.DownloadingFile})
	if err := util.DownloadFile(filePath, message.CommandArguments()); err != nil {
		telegram.SendMessage(telegram.Message{ChatID: message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorDownloadFileFailed, err.Error())})
		return
	}

	buildExtract(filePath, "extracturl", message)
}

// ExtractChoose handles the user choice of which file(s) to extract from the current session archive
func ExtractChoose(callbackQuery *tgbotapi.CallbackQuery, callbackData callback.Callback) {
	session, exists := session.Session(callbackQuery.Message.Chat.ID)
	if !exists {
		telegram.SendMessage(telegram.Message{ChatID: callbackQuery.Message.Chat.ID, Text: i18n.ErrorNoSession})
		return
	}

	i := 0
	err := archiver.Walk(session.ActionFile, func(f archiver.File) error {
		if (callbackData.HasIndex(i) || callbackData.HasFlag(callback.FlagWildcard)) && f.Size() <= commons.Config.MaxSingleFileSize && f.Size() > 0 {
			telegram.SendMessage(telegram.Message{ChatID: callbackQuery.Message.Chat.ID, Text: i18n.SendingFile})
			if err := telegram.SendFile(callbackQuery.Message.Chat.ID, tgbotapi.FileReader{Name: f.Name(), Size: f.Size(), Reader: f.ReadCloser}); err != nil {
				log.Printf("ExtractChoose: SendFile: %s\n", err)
			}
		}
		i++
		return nil
	})
	if err != nil {
		telegram.SendMessage(telegram.Message{ChatID: callbackQuery.Message.Chat.ID, Text: fmt.Sprintf("%s: %s", i18n.ErrorArchiveWalkFailed, err.Error())})
		return
	}
}

// ExtractDone ends the current extract session, if any
func ExtractDone(callbackQuery *tgbotapi.CallbackQuery) {
	session, exists := session.Session(callbackQuery.Message.Chat.ID)
	if !exists {
		telegram.SendMessage(telegram.Message{ChatID: callbackQuery.Message.Chat.ID, Text: i18n.ErrorNoSession})
		return
	}
	defer session.Delete()

	if err := telegram.ClearKeyboardMarkup(callbackQuery.Message); err != nil {
		log.Printf("ExtractDone: ClearKeyboardMarkup: %s\n", err)
	}

	telegram.SendMessage(telegram.Message{ChatID: callbackQuery.Message.Chat.ID, Text: i18n.SessionOver})
}
