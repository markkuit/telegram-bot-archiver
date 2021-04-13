package i18n

const (
	DownloadingFile      = "Downloading file.."
	SendingFile          = "Sending file.."
	SessionAlreadyActive = "A session is already active. You can force its deletion with `/cancel`."
	SessionDeleted       = "Session has been deleted."
	SessionOver          = "Session over. Thank you for using this bot!"
	Usage                = `Welcome to the Telegram Archiver bot!

Archiver is capable of extracting archives from the web or Telegram documents, as well as creating new ones from your own files.
This bot has been developed by @markkuit and it is [open source](https://github.com/markkuit/telegram-bot-archiver).

Type ` + "`" + `/extractfile` + "`" + ` to upload an archive to be extracted
Type ` + "`" + `/extracturl <URL>` + "`" + ` to provide a _direct link_ to an archive to be extracted, e.g. ` + "`" + `/extracturl https://markkuit.net/myDocuments.zip` + "`" + `
Type ` + "`" + `/compress <filename.ext>` + "`" + ` to create a new archive, e.g. ` + "`" + `/compress myDocuments.zip` + "`" + `

Get in touch if you need help. Enjoy!
	`

	CompressInitializationFailed = "Compress initialization failed. The specified archive format may be unsupported."
	CompressProcessedFile        = "File has been processed"
	CompressSendFiles            = "Session active. Send files to be compressed. Type `/compressdone` when finished."

	ExtractChooseFiles             = "Now listing suitable archive files; go ahead and choose."
	ExtractMaxInlineButtonsReached = "NOTE: the maximum number of listable files has been reached, listing will be partial. Get all files if necessary."
	ExtractMaxFilesReached         = "NOTE: the maximum number of files to massively get has been reached."
	ExtractNoValidFiles            = "No suitable files could be find in the provided archive."
	ExtractSendArchive             = "Session active. Send archive to be extracted."

	ErrorArchiveWalkFailed         = "Error while iterating through archive files"
	ErrorArgumentMissing           = "Argument missing. Type `/start` for help."
	ErrorDownloadFileFailed        = "Error downloading file"
	ErrorGetFileFailed             = "Error getting file"
	ErrorHTTPHeadFailed            = "Error during preliminary HTTP requests"
	ErrorHTTPInvalidContent        = "Invalid content on HTTP request"
	ErrorMaxContentLengthExceeded  = "Maximum content size exceeded"
	ErrorMaxSingleFileSizeExceeded = "Maximum allowed single file size exceeded"
	ErrorNoSession                 = "No active session could be found."
	ErrorUnexpectedDocument        = "Unexpected document received. Type `/start` for help."
	ErrorUnrecognizedCommand       = "Unrecognized command received. Type `/start` for help."
	ErrorURLParseFailed            = "Error while parsing URL"
)
