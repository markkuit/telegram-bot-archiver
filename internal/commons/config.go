package commons

import (
	"log"
	"os"
	"strconv"

	"github.com/dustin/go-humanize"
	_ "github.com/joho/godotenv/autoload"
)

type botConfig struct {
	APIToken             string
	MaxContentLength     int64
	MaxExtractFiles      int
	MaxInlineButtons     int
	MaxSingleFileSize    int64
	InlineButtonsRowSize int
	TempPath             string
}

// Config holds the bot configuration
var Config botConfig

func init() {
	Config.APIToken = getenv("API_TOKEN", true)
	Config.MaxContentLength = parseBytes(getenv("MAX_CONTENT_LENGTH", true))
	Config.MaxExtractFiles, _ = strconv.Atoi(getenv("MAX_EXTRACT_FILES", true))
	Config.MaxInlineButtons, _ = strconv.Atoi(getenv("MAX_INLINE_BUTTONS", true))
	Config.MaxSingleFileSize = parseBytes(getenv("MAX_SINGLE_FILE_SIZE", true))
	Config.InlineButtonsRowSize, _ = strconv.Atoi(getenv("INLINE_BUTTONS_ROW_SIZE", true))
	Config.TempPath = getenv("TEMP_PATH", true)

	if _, err := os.Stat(Config.TempPath); os.IsNotExist(err) {
		if err := os.Mkdir(Config.TempPath, os.ModePerm); err != nil {
			log.Fatalf("init: Mkdir: %s\n", err)
		}
	}

	if Config.InlineButtonsRowSize > 8 {
		log.Fatal("INLINE_BUTTONS_ROW_SIZE: Telegram Bot API allows no more than 8 buttons per KeyboardButtonRow")
	}
}

// parseBytes wraps humanize.ParseBytes() for error handling and int64 casting.
// Although taking away error handling from context is generally a bad practice,
// we already know for sure what use we're going to make of this func
func parseBytes(s string) int64 {
	bytes, err := humanize.ParseBytes(s)
	if err != nil {
		log.Fatalf("parseBytes: %s\n", err)
	}
	return int64(bytes)
}

// getenv wraps the env lookup for less boilerplate and existence checking
func getenv(key string, mandatory bool) string {
	value, ok := os.LookupEnv(key)
	if !ok && mandatory {
		log.Fatalf("getenv: env var %s is mandatory\n", key)
	}
	return value
}
