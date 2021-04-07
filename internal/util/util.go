package util

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/markkuit/telegram-bot-archiver/internal/commons"
	"github.com/markkuit/telegram-bot-archiver/internal/i18n"
	"github.com/mholt/archiver"
)

// ArchiveWalk walks through the archive files, skipping those bigger
// than configured and returning an indexed list
func ArchiveWalk(filePath string) (map[int]archiver.File, error) {
	var files = make(map[int]archiver.File)
	i := 0
	if err := archiver.Walk(filePath, func(f archiver.File) error {
		if f.Size() <= commons.Config.MaxSingleFileSize {
			files[i] = f
		}
		i++
		return nil
	}); err != nil {
		return map[int]archiver.File{}, err
	}
	return files, nil
}

// DownloadFile takes a URL and downloads it to the given path. It abides by the configured limits.
func DownloadFile(dst string, src string) error {
	srcURL, err := url.Parse(src)
	if err != nil {
		return fmt.Errorf("%s: %s", i18n.ErrorURLParseFailed, err.Error())
	}

	res, err := http.Head(srcURL.String())
	if err != nil {
		return fmt.Errorf("%s: %s", i18n.ErrorHTTPHeadFailed, err.Error())
	}
	if res.ContentLength <= 0 {
		return errors.New("SendFile: " + i18n.ErrorHTTPInvalidContent)
	}
	if res.ContentLength > commons.Config.MaxContentLength {
		return errors.New("DownloadFile: " + i18n.ErrorMaxContentLengthExceeded)
	}

	res, err = http.Get(srcURL.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if _, err := os.Stat(filepath.Dir(dst)); os.IsNotExist(err) {
		if err := os.Mkdir(filepath.Dir(dst), os.ModePerm); err != nil {
			log.Fatalf("DownloadFile: Mkdir: %s\n", err)
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return err
}
