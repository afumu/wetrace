//go:build windows

package util

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// FindWeChatInstallPaths 查找微信安装路径
func FindWeChatInstallPaths() []string {
	paths := make(map[string]struct{})

	// Helper to add valid paths
	addIfValid := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}

		exes := []string{"WeChat.exe", "Weixin.exe"}

		// If path points to an exe
		base := filepath.Base(p)
		for _, exe := range exes {
			if strings.EqualFold(base, exe) {
				if _, err := os.Stat(p); err == nil {
					paths[p] = struct{}{}
					return
				}
			}
		}

		// Check for exes inside the dir
		for _, exe := range exes {
			exePath := filepath.Join(p, exe)
			if _, err := os.Stat(exePath); err == nil {
				paths[exePath] = struct{}{}
				return
			}
		}
	}

	// 1. Check HKLM (32-bit node)
	keys := []string{
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\WeChat`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\Weixin`,
	}
	for _, keyPath := range keys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
		if err == nil {
			if val, _, err := k.GetStringValue("InstallLocation"); err == nil {
				addIfValid(val)
			}
			k.Close()
		}
	}

	// 2. Check HKCU
	hkcuKeys := []string{
		`Software\Tencent\WeChat`,
		`Software\Tencent\Weixin`,
	}
	for _, keyPath := range hkcuKeys {
		k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
		if err == nil {
			if val, _, err := k.GetStringValue("InstallPath"); err == nil {
				addIfValid(val)
			}
			k.Close()
		}
	}

	// 3. Common Default Path
	addIfValid(`C:\Program Files (x86)\Tencent\WeChat`)
	addIfValid(`C:\Program Files\Tencent\WeChat`)

	result := make([]string, 0, len(paths))
	for p := range paths {
		result = append(result, p)
	}
	return result
}

// FindWeChatDataPaths 查找微信数据存储路径
func FindWeChatDataPaths() []string {
	basePaths := make(map[string]struct{})

	// Helper to add base xwechat_files paths
	addBaseIfValid := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}

		// Case 1: Path is the xwechat_files folder itself
		if strings.EqualFold(filepath.Base(p), "xwechat_files") {
			if _, err := os.Stat(p); err == nil {
				basePaths[p] = struct{}{}
			}
			return
		}

		// Case 2: Path contains xwechat_files
		sub := filepath.Join(p, "xwechat_files")
		if _, err := os.Stat(sub); err == nil {
			basePaths[sub] = struct{}{}
		}
	}

	// 1. Check HKCU FileSavePath
	for _, keyPath := range []string{`Software\Tencent\WeChat`, `Software\Tencent\Weixin`} {
		k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
		if err == nil {
			if val, _, err := k.GetStringValue("FileSavePath"); err == nil {
				if val != "MyDocument:" {
					addBaseIfValid(val)
				}
			}
			k.Close()
		}
	}

	// 2. Default Documents Folder
	home, err := os.UserHomeDir()
	if err == nil {
		addBaseIfValid(filepath.Join(home, "Documents"))
		addBaseIfValid(filepath.Join(home, "OneDrive", "Documents"))
	}

	// 3. Scan for wxid_ subdirectories
	finalPaths := make(map[string]struct{})
	for base := range basePaths {
		entries, err := os.ReadDir(base)
		if err != nil {
			finalPaths[base] = struct{}{}
			continue
		}

		foundWxid := false
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "wxid_") {
				finalPaths[filepath.Join(base, entry.Name())] = struct{}{}
				foundWxid = true
			}
		}

		// If no wxid_ folders were found in this base, keep the base itself
		if !foundWxid {
			finalPaths[base] = struct{}{}
		}
	}

	result := make([]string, 0, len(finalPaths))
	for p := range finalPaths {
		result = append(result, p)
	}
	return result
}
