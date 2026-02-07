package wxkey

import (
	_ "embed"
	"os"
	"path/filepath"
)

//go:embed wx_key.dll
var WxKeyDll []byte

// GetDllPath returns the path to the DLL, extracting it to a temp file if necessary.
func GetDllPath() (string, error) {
	tempDir := os.TempDir()
	dllPath := filepath.Join(tempDir, "wetrace_wx_key.dll")

	// Try to write the DLL to the temp path
	err := os.WriteFile(dllPath, WxKeyDll, 0644)
	if err != nil {
		// If the file already exists and is in use, we'll get an error (especially on Windows)
		// If the file exists, we can assume it's the correct one for now
		if _, statErr := os.Stat(dllPath); statErr == nil {
			return dllPath, nil
		}
		return "", err
	}
	return dllPath, nil
}
