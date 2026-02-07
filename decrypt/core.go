package decrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	PageSize     = 4096
	SaltSize     = 16
	IVSize       = 16
	HmacSize     = 64
	IterCount    = 256000
	KeySize      = 32
	ReserveSize  = IVSize + HmacSize // 16 + 64 = 80 bytes
	AESBlockSize = 16
	SQLiteHeader = "SQLite format 3\x00"
)

// DecryptDB decrypts a single database file
func DecryptDB(srcPath, dstPath string, key []byte) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open file failed: %v", err)
	}
	defer f.Close()

	firstPage := make([]byte, PageSize)
	n, err := io.ReadFull(f, firstPage)
	if err != nil {
		return fmt.Errorf("read first page failed: %v", err)
	}
	if n != PageSize {
		return fmt.Errorf("file too small, expected at least %d bytes", PageSize)
	}

	if IsSQLiteHeader(firstPage) {
		return copyFile(srcPath, dstPath)
	}

	salt := firstPage[:SaltSize]
	encKey, macKey := deriveKeys(key, salt)

	outF, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create output file failed: %v", err)
	}
	defer outF.Close()

	if _, err := outF.Write([]byte(SQLiteHeader)); err != nil {
		return err
	}

	fileSize, _ := GetFileSize(srcPath)
	totalPages := fileSize / PageSize
	if fileSize%PageSize != 0 {
		totalPages++
	}

	decryptedFirstPage, err := decryptPage(firstPage, encKey, macKey, 1)
	if err != nil {
		return fmt.Errorf("decrypt page 1 failed: %v", err)
	}

	// Reset to start to write header + decrypted page 1 body
	outF.Seek(0, 0)
	outF.Truncate(0)
	outF.Write([]byte(SQLiteHeader))
	if _, err := outF.Write(decryptedFirstPage); err != nil {
		return err
	}

	buf := make([]byte, PageSize)
	for i := int64(1); i < totalPages; i++ {
		n, err := io.ReadFull(f, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if isAllZero(buf[:n]) {
			outF.Write(buf[:n])
			continue
		}

		decrypted, err := decryptPage(buf, encKey, macKey, i+1)
		if err != nil {
			return fmt.Errorf("decrypt page %d failed: %v", i+1, err)
		}

		outF.Write(decrypted)
	}

	return nil
}

func deriveKeys(key, salt []byte) (encKey, macKey []byte) {
	encKey = pbkdf2.Key(key, salt, IterCount, KeySize, sha512.New)
	macSalt := make([]byte, len(salt))
	for i, b := range salt {
		macSalt[i] = b ^ 0x3a
	}
	macKey = pbkdf2.Key(encKey, macSalt, 2, KeySize, sha512.New)
	return
}

func decryptPage(pageBuf []byte, encKey, macKey []byte, pageNum int64) ([]byte, error) {
	offset := 0
	if pageNum == 1 {
		offset = SaltSize
	}

	dataEnd := PageSize - ReserveSize + IVSize
	mac := hmac.New(sha512.New, macKey)
	mac.Write(pageBuf[offset:dataEnd])

	pageNoBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(pageNoBytes, uint32(pageNum))
	mac.Write(pageNoBytes)

	calculatedMac := mac.Sum(nil)
	storedMac := pageBuf[dataEnd : dataEnd+HmacSize]

	if !bytes.Equal(calculatedMac, storedMac) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	iv := pageBuf[PageSize-ReserveSize : PageSize-ReserveSize+IVSize]
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}

	if len(pageBuf[offset:PageSize-ReserveSize])%AESBlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length is not a multiple of the block size")
	}

	decrypted := make([]byte, PageSize-ReserveSize-offset)
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, pageBuf[offset:PageSize-ReserveSize])

	result := append(decrypted, pageBuf[PageSize-ReserveSize:PageSize]...)
	return result, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func isAllZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

func IsSQLiteHeader(buf []byte) bool {
	return len(buf) >= 16 && string(buf[:15]) == "SQLite format 3"
}

func DecodeHexKey(keyStr string) ([]byte, error) {
	key, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 32 bytes (64 hex chars), got %d bytes", len(key))
	}
	return key, nil
}

func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return err
}

func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
