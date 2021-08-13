package logger

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"time"
)

// Logger is custom console logger
type Logger struct {
}

// Write writing bytes to log
func (writer Logger) Write(bytes []byte) (int, error) {
	return fmt.Print("[APP] " + time.Now().UTC().Format("2006/01/02 - 15:04:05") + " " + string(bytes))
}

// LogSuccess logs success message
func LogSuccess(message string, prefix string) {
	var str = ""
	if prefix != "" {
		color := color.New(color.FgHiWhite).Add(color.BgGreen)
		str += "|" + color.Sprint(prefix) + "| "
	}
	str += message
	log.Print(str)
}

// LogWarning logs warning message
func LogWarning(message string, prefix string) {
	var str = ""
	if prefix != "" {
		color := color.New(color.FgHiWhite).Add(color.BgYellow)
		str += "|" + color.Sprint(prefix) + "| "
	}
	str += message
	log.Print(str)
}

// LogError logs error message
func LogError(message string, prefix string) {
	var str = ""
	if prefix != "" {
		color := color.New(color.FgHiWhite).Add(color.BgRed)
		str += "|" + color.Sprint(prefix) + "| "
	}
	str += message
	log.Print(str)
}
