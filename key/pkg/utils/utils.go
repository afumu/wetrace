package utils

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")

	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
)

const (
	SW_HIDE = 0
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func HideConsoleWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		procShowWindow.Call(hwnd, SW_HIDE)
	}
}

func GetExecutableDirectory() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(ex), nil
}

func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
