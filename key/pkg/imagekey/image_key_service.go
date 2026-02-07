package imagekey

import (
	"crypto/aes"
	_ "encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/afumu/wetrace/key/pkg/logger"
	"github.com/afumu/wetrace/key/pkg/process"

	"golang.org/x/sys/windows"
)

type ImageKeyResult struct {
	XorKey  int
	AesKey  string
	Success bool
	Error   error
}

type ImageKeyService struct {
	logger *logger.Logger
}

func NewImageKeyService(log *logger.Logger) *ImageKeyService {
	return &ImageKeyService{
		logger: log,
	}
}

func (s *ImageKeyService) GetImageKeys(pid uint32, manualDataPath string) ImageKeyResult {
	s.logger.Info("开始获取图片密钥")

	// 1. Locate WeChat Cache Directory
	s.logger.Info("正在定位微信缓存目录...")

	var cacheDir string
	var err error

	if manualDataPath != "" {
		s.logger.Info("使用手动指定的路径: " + manualDataPath)
		// Check if it is the account dir directly
		if s.directoryHasDbStorage(manualDataPath) || s.directoryHasImageCache(manualDataPath) {
			cacheDir = manualDataPath
		} else {
			// Try to scan as root
			cacheDir, err = s.scanForAccountDir(manualDataPath)
		}
	} else {
		cacheDir, err = s.autoLocateWeChatCacheDirectory()
	}

	if err != nil || cacheDir == "" {
		s.logger.Error("未找到微信缓存目录")
		return ImageKeyResult{Error: fmt.Errorf("未找到微信缓存目录")}
	}
	s.logger.Info("找到缓存目录: " + cacheDir)

	// 2. Find template files
	s.logger.Info("正在收集模板文件...")
	templates, err := s.findTemplateDatFiles(cacheDir)
	if err != nil || len(templates) == 0 {
		s.logger.Error("未找到模板文件")
		return ImageKeyResult{Error: fmt.Errorf("未找到模板文件")}
	}
	s.logger.Info(fmt.Sprintf("找到 %d 个模板文件", len(templates)))

	// 3. Calculate XOR Key
	s.logger.Info("正在计算 XOR 密钥...")
	xorKey, err := s.getXorKey(templates)
	if err != nil {
		s.logger.Error("无法获取 XOR 密钥")
		return ImageKeyResult{Error: err}
	}
	s.logger.Info(fmt.Sprintf("成功获取 XOR 密钥: %02X", xorKey))

	// 4. Get Ciphertext
	s.logger.Info("正在读取加密数据...")
	ciphertext, err := s.getCiphertextFromTemplate(templates)
	if err != nil {
		s.logger.Error("无法读取加密数据")
		return ImageKeyResult{Error: err}
	}
	s.logger.Info(fmt.Sprintf("成功读取 %d 字节加密数据", len(ciphertext)))

	// 5. Scan Memory for AES Key
	s.logger.Info(fmt.Sprintf("开始从内存中搜索 AES 密钥 (PID: %d)...", pid))
	aesKey, err := s.scanMemoryForAesKey(pid, ciphertext)
	if err != nil {
		s.logger.Error("内存搜索失败: " + err.Error())
		return ImageKeyResult{Error: err}
	}

	s.logger.Success("成功获取 AES 密钥: " + aesKey)
	s.logger.Success("图片密钥获取完成")

	return ImageKeyResult{
		XorKey:  xorKey,
		AesKey:  aesKey,
		Success: true,
	}
}

// ----------------------------------------------------------------------------
// Step 1: Find Cache Directory
// ----------------------------------------------------------------------------

func (s *ImageKeyService) autoLocateWeChatCacheDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	wechatFilesPath := filepath.Join(homeDir, "Documents", "xwechat_files")
	return s.scanForAccountDir(wechatFilesPath)
}

func (s *ImageKeyService) scanForAccountDir(rootPath string) (string, error) {
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return "", nil
	}

	entries, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return "", err
	}

	var highConfidence []string
	var lowConfidence []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !s.isPotentialAccountDirectory(name) {
			continue
		}

		fullPath := filepath.Join(rootPath, name)
		hasDb := s.directoryHasDbStorage(fullPath)
		hasImg := s.directoryHasImageCache(fullPath)

		if hasDb || hasImg {
			highConfidence = append(highConfidence, fullPath)
		} else {
			lowConfidence = append(lowConfidence, fullPath)
		}
	}

	if len(highConfidence) > 0 {
		// Sort to pick 'latest' or alphabetical? Dart sorts alphabetically by basename
		sort.Slice(highConfidence, func(i, j int) bool {
			return filepath.Base(highConfidence[i]) < filepath.Base(highConfidence[j])
		})
		// Picking the first one is a heuristic. In multi-account scenarios, this might be wrong.
		// Ideally we match the PID user, but that's hard.
		return highConfidence[0], nil
	}

	if len(lowConfidence) > 0 {
		sort.Slice(lowConfidence, func(i, j int) bool {
			return filepath.Base(lowConfidence[i]) < filepath.Base(lowConfidence[j])
		})
		return lowConfidence[0], nil
	}

	return "", nil
}

func (s *ImageKeyService) isPotentialAccountDirectory(dirName string) bool {
	lower := strings.ToLower(dirName)
	if strings.HasPrefix(lower, "all") ||
		strings.HasPrefix(lower, "applet") ||
		strings.HasPrefix(lower, "backup") ||
		strings.HasPrefix(lower, "wmpf") {
		return false
	}
	return strings.HasPrefix(dirName, "wxid_") || len(dirName) > 5
}

func (s *ImageKeyService) directoryHasDbStorage(path string) bool {
	info, err := os.Stat(filepath.Join(path, "db_storage"))
	return err == nil && info.IsDir()
}

func (s *ImageKeyService) directoryHasImageCache(path string) bool {
	info, err := os.Stat(filepath.Join(path, "FileStorage", "Image"))
	return err == nil && info.IsDir()
}

// ----------------------------------------------------------------------------
// Step 2: Find Template Files
// ----------------------------------------------------------------------------

func (s *ImageKeyService) findTemplateDatFiles(userDir string) ([]string, error) {
	var files []string
	maxFiles := 32

	// Walk dir
	err := filepath.Walk(userDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore errors
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_t.dat") {
			files = append(files, path)
			if len(files) >= maxFiles {
				// return error to stop walk? No, just keep going or customize Walk.
				// Standard Walk doesn't support 'Stop' easily without error.
				// Let's just collect up to maxFiles + some and sort.
			}
		}
		return nil
	})

	if len(files) == 0 {
		return nil, err
	}

	// Sort by date from path (regex) desc
	// Pattern: (\d{4}-\d{2})
	re := regexp.MustCompile(`(\d{4}-\d{2})`)

	sort.Slice(files, func(i, j int) bool {
		matchA := re.FindString(files[i])
		matchB := re.FindString(files[j])
		// Descending
		return matchB < matchA
	})

	if len(files) > 16 {
		files = files[:16]
	}
	return files, nil
}

// ----------------------------------------------------------------------------
// Step 3: Get XOR Key
// ----------------------------------------------------------------------------

func (s *ImageKeyService) getXorKey(files []string) (int, error) {
	lastBytesMap := make(map[string]int)

	for _, fpath := range files {
		content, err := ioutil.ReadFile(fpath)
		if err != nil || len(content) < 2 {
			continue
		}
		lastTwo := content[len(content)-2:]
		key := fmt.Sprintf("%d_%d", lastTwo[0], lastTwo[1])
		lastBytesMap[key]++
	}

	if len(lastBytesMap) == 0 {
		return 0, fmt.Errorf("no valid files for xor calculation")
	}

	var mostCommon string
	var maxCount int
	for k, v := range lastBytesMap {
		if v > maxCount {
			maxCount = v
			mostCommon = k
		}
	}

	if mostCommon != "" {
		parts := strings.Split(mostCommon, "_")
		x, _ := strconv.Atoi(parts[0])
		y, _ := strconv.Atoi(parts[1])

		xorKey := x ^ 0xFF
		check := y ^ 0xD9

		if xorKey == check {
			return xorKey, nil
		}
	}
	return 0, fmt.Errorf("failed to calculate xor key")
}

// ----------------------------------------------------------------------------
// Step 4: Get Ciphertext
// ----------------------------------------------------------------------------

func (s *ImageKeyService) getCiphertextFromTemplate(files []string) ([]byte, error) {
	// Header signature: 07 08 56 32 08 07
	sig := []byte{0x07, 0x08, 0x56, 0x32, 0x08, 0x07}

	for _, fpath := range files {
		content, err := ioutil.ReadFile(fpath)
		if err != nil || len(content) < 0x1F {
			continue
		}

		if len(content) >= 6 && equalBytes(content[:6], sig) {
			// Ciphertext is at 0xF to 0x1F (16 bytes)
			return content[0xF:0x1F], nil
		}
	}
	return nil, fmt.Errorf("no matching template file found")
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------------
// Step 5: Scan Memory
// ----------------------------------------------------------------------------

func (s *ImageKeyService) scanMemoryForAesKey(pid uint32, ciphertext []byte) (string, error) {
	hProcess, err := process.OpenProcess(pid)
	if err != nil {
		return "", fmt.Errorf("OpenProcess failed: %v", err)
	}
	defer windows.CloseHandle(hProcess)

	regions, err := process.GetMemoryRegions(hProcess)
	if err != nil {
		return "", fmt.Errorf("GetMemoryRegions failed: %v", err)
	}

	s.logger.Info(fmt.Sprintf("扫描内存区域: %d 个", len(regions)))

	scanned := 0
	skipped := 0

	for i, region := range regions {
		// Skip large regions
		if region.RegionSize > 100*1024*1024 {
			skipped++
			continue
		}

		scanned++

		// Read Memory
		data, err := process.ReadProcessMemory(hProcess, region.BaseAddress, region.RegionSize)
		if err != nil || len(data) == 0 {
			continue
		}

		// Search for pattern: [not a-z0-9] [a-z0-9]{32} [not a-z0-9]
		for j := 0; j < len(data)-34; j++ {
			// Check leading char (must NOT be lower alphanumeric)
			if s.isAlphaNumLower(data[j]) {
				continue
			}

			// Check trailing char (must NOT be lower alphanumeric)
			if s.isAlphaNumLower(data[j+33]) {
				continue
			}

			// Check the 32 chars in between
			candidate := data[j+1 : j+33]
			isValid := true
			for k := 0; k < 32; k++ {
				if !s.isAlphaNumLower(candidate[k]) {
					isValid = false
					break
				}
			}

			if isValid {
				// Verify
				keyStr := string(candidate)
				if s.verifyKey(ciphertext, keyStr) {
					s.logger.Success(fmt.Sprintf("在第 %d 个区域找到 AES 密钥 (基址: 0x%X)", i, region.BaseAddress))
					return keyStr[:16], nil
				}
			}
		}
	}

	s.logger.Info(fmt.Sprintf("扫描结束。已扫描区域: %d, 跳过(过大): %d, 未找到匹配项", scanned, skipped))
	return "", fmt.Errorf("key not found in memory")
}

func (s *ImageKeyService) isAlphaNumLower(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

// ----------------------------------------------------------------------------
// Step 6: Verify Key
// ----------------------------------------------------------------------------

func (s *ImageKeyService) verifyKey(encrypted []byte, aesKeyStr string) bool {
	// aesKeyStr is 32 chars hex string? No, memory scan finds 32 chars of "lowercase alphanumeric".
	// The Dart code says: `aesKey.sublist(0, 16)`.
	// Wait, the memory pattern is 32 bytes of ASCII chars.
	// `key = aesKey.sublist(0, 16)` means taking the first 16 bytes of that ASCII string as the key bytes?
	// Yes, `AES.new(key, AES.MODE_ECB)`.

	if len(aesKeyStr) < 16 {
		return false
	}

	keyBytes := []byte(aesKeyStr[:16])

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return false
	}

	// ECB Decrypt
	// We only need to decrypt the first block to check signature
	if len(encrypted) < block.BlockSize() {
		return false
	}

	decrypted := make([]byte, len(encrypted))

	// Process block by block
	for i := 0; i < len(encrypted); i += block.BlockSize() {
		if i+block.BlockSize() > len(encrypted) {
			break
		}
		block.Decrypt(decrypted[i:i+block.BlockSize()], encrypted[i:i+block.BlockSize()])
	}

	// Check header: FF D8 FF
	if len(decrypted) >= 3 &&
		decrypted[0] == 0xFF &&
		decrypted[1] == 0xD8 &&
		decrypted[2] == 0xFF {
		return true
	}

	return false
}
