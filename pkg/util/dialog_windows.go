//go:build windows

package util

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"
)

// OpenFileDialog 打开文件选择对话框
func OpenFileDialog(title string, filter string) (string, error) {
	// PowerShell script to open FileDialog
	psScript := `
	Add-Type -AssemblyName System.Windows.Forms
	$f = New-Object System.Windows.Forms.OpenFileDialog
	$f.Title = "` + title + `"
	$f.Filter = "` + filter + `"
	$f.ShowHelp = $true
	if ($f.ShowDialog() -eq "OK") {
		return $f.FileName
	} else {
		return "CANCELLED"
	}
	`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // Hide PowerShell window
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(out))
	if result == "CANCELLED" || result == "" {
		return "", errors.New("cancelled")
	}

	return result, nil
}

// OpenFolderDialog 打开文件夹选择对话框
func OpenFolderDialog(description string) (string, error) {
	// PowerShell script to open FolderBrowserDialog
	// Note: System.Windows.Forms.FolderBrowserDialog is a bit old style but works reliably.
	psScript := `
	Add-Type -AssemblyName System.Windows.Forms
	$f = New-Object System.Windows.Forms.FolderBrowserDialog
	$f.Description = "` + description + `"
	$f.ShowNewFolderButton = $false
	if ($f.ShowDialog() -eq "OK") {
		return $f.SelectedPath
	} else {
		return "CANCELLED"
	}
	`

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(out))
	if result == "CANCELLED" || result == "" {
		return "", errors.New("cancelled")
	}

	return result, nil
}
