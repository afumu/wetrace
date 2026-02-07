package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	Verbose bool
	Quiet   bool
	MuteAll bool
	NoColor bool
	LogFile *os.File
}

func NewLogger(verbose, quiet, noColor bool) *Logger {
	return &Logger{
		Verbose: verbose,
		Quiet:   quiet,
		NoColor: noColor,
	}
}

func (l *Logger) SetLogFile(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	l.LogFile = f
	return nil
}

func (l *Logger) Close() {
	if l.LogFile != nil {
		l.LogFile.Close()
	}
}

func (l *Logger) log(level, msg string, colorCode string) {
	if l.MuteAll {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMsg := fmt.Sprintf("[%s] %s", level, msg)

	if !l.Quiet || level == "ERROR" {
		if !l.NoColor && colorCode != "" {
			fmt.Printf("%s%s\033[0m\n", colorCode, formattedMsg)
		} else {
			fmt.Println(formattedMsg)
		}
	}

	if l.LogFile != nil {
		l.LogFile.WriteString(fmt.Sprintf("%s %s\n", timestamp, formattedMsg))
	}
}

func (l *Logger) Info(msg string) {
	l.log("INFO", msg, "\033[36m") // Cyan
}

func (l *Logger) Success(msg string) {
	l.log("SUCCESS", msg, "\033[32m") // Green
}

func (l *Logger) Warning(msg string) {
	l.log("WARNING", msg, "\033[33m") // Yellow
}

func (l *Logger) Error(msg string) {
	l.log("ERROR", msg, "\033[31m") // Red
}

func (l *Logger) Debug(msg string) {
	if l.Verbose {
		l.log("DEBUG", msg, "\033[90m") // Gray
	}
}

func (l *Logger) LogRaw(msg string) {
	if !l.Quiet {
		fmt.Println(msg)
	}
	if l.LogFile != nil {
		l.LogFile.WriteString(msg + "\n")
	}
}

func (l *Logger) LogToFile(msg string) {
	if l.LogFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		l.LogFile.WriteString(fmt.Sprintf("%s [RESULT] %s\n", timestamp, msg))
	}
}
