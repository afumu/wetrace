//go:build !windows

package util

import "errors"

func OpenFileDialog(title string, filter string) (string, error) {
	return "", errors.New("not supported on this OS")
}

func OpenFolderDialog(description string) (string, error) {
	return "", errors.New("not supported on this OS")
}
