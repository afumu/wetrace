package decrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RunTask executes the decryption process with provided parameters.
func RunTask(srcDir, keyStr string) (int, string, error) {
	if keyStr == "" {
		return 0, "", fmt.Errorf("数据库密钥未获取,请先获取密钥")
	}

	if srcDir == "" {
		return 0, "", fmt.Errorf("微信存储路径未配置，请先配置存储路径")
	}

	// 检查 srcDir 是否存在且为目录
	stat, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", fmt.Errorf("指定的微信存储路径不存在: %s", srcDir)
		}
		return 0, "", fmt.Errorf("无法访问微信存储路径: %v", err)
	}
	if !stat.IsDir() {
		return 0, "", fmt.Errorf("指定的微信存储路径不是一个目录: %s", srcDir)
	}

	// 自动拼接 db_storage 目录进行扫描
	actualSrcDir := filepath.Join(srcDir, "db_storage")

	key, err := DecodeHexKey(keyStr)
	if err != nil {
		return 0, "", fmt.Errorf("invalid key: %v", err)
	}

	outDir := "data"
	if err := EnsureDir(outDir); err != nil {
		return 0, "", fmt.Errorf("create output dir failed: %v", err)
	}

	var dbFiles []string
	err = filepath.Walk(actualSrcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 如果 db_storage 不存在，尝试直接扫描 srcDir
			if os.IsNotExist(err) && path == actualSrcDir {
				return nil
			}
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".db") {
			if strings.Contains(strings.ToLower(info.Name()), "fts") {
				return nil
			}
			dbFiles = append(dbFiles, path)
		}
		return nil
	})

	// 如果在 db_storage 没找到，尝试回退到原始目录扫描
	if len(dbFiles) == 0 {
		_ = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".db") {
				if !strings.Contains(strings.ToLower(info.Name()), "fts") {
					dbFiles = append(dbFiles, path)
				}
			}
			return nil
		})
	}

	if len(dbFiles) == 0 {
		return 0, "", fmt.Errorf("在指定目录及其子目录下未找到任何微信数据库文件(.db)，请检查路径是否正确")
	}

	if err != nil {
		return 0, "", fmt.Errorf("扫描目录失败: %v", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	successCount := 0
	var mu sync.Mutex
	var firstErr error

	for _, src := range dbFiles {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 优先尝试相对于 actualSrcDir 的路径，如果不行则尝试相对于 srcDir
			rel, err := filepath.Rel(actualSrcDir, src)
			if err != nil || strings.HasPrefix(rel, "..") {
				rel, _ = filepath.Rel(srcDir, src)
			}

			if rel == "" || rel == "." {
				rel = filepath.Base(src)
			}
			dst := filepath.Join(outDir, rel)

			if err := EnsureDir(filepath.Dir(dst)); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			if err := DecryptDB(src, dst, key); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				fmt.Printf("Failed to decrypt %s: %v\n", rel, err)
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(src)
	}

	wg.Wait()

	if successCount == 0 && len(dbFiles) > 0 {
		return 0, outDir, fmt.Errorf("failed to decrypt any files (found %d). First error: %v", len(dbFiles), firstErr)
	}

	return successCount, outDir, nil
}

func loadEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return nil
}
