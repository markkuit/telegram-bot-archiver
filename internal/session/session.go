package session

import (
	"log"
	"os"
)

var sessions map[int64]session

type session struct {
	ChatID       int64
	Action       string
	ActionFile   string
	SessionFiles []string
}

func init() {
	sessions = make(map[int64]session)
}

// New creates a new session for the given chatID
func New(chatID int64, action, actionFile string) session {
	var s = session{
		ChatID:     chatID,
		Action:     action,
		ActionFile: actionFile,
	}
	sessions[chatID] = s
	return s
}

// Session returns the given chatID associated session, if any
func Session(chatID int64) (session, bool) {
	if !Exists(chatID) {
		return session{}, false
	}
	return sessions[chatID], true
}

// Exists checks whether the given chatID has an associated session
func Exists(chatID int64) bool {
	if _, ok := sessions[chatID]; ok {
		return true
	}
	return false
}

// Delete deletes the session
func (s session) Delete() {
	s.cleanup()
	delete(sessions, s.ChatID)
}

// AddSessionFile adds a file to the given session
func (s session) AddSessionFile(filePath string) {
	// session := sessions[s.ChatID]
	s.SessionFiles = append(s.SessionFiles, filePath)
	sessions[s.ChatID] = s
}

func (s session) cleanup() {
	for _, tempFile := range append(s.SessionFiles, s.ActionFile) {
		if err := os.RemoveAll(tempFile); err != nil {
			log.Printf("cleanup: RemoveAll: %s\n", err)
		}
	}
}
