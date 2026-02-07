package appver

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	modversion                 = syscall.NewLazyDLL("version.dll")
	procGetFileVersionInfoSize = modversion.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfo     = modversion.NewProc("GetFileVersionInfoW")
	procVerQueryValue          = modversion.NewProc("VerQueryValueW")
)

// VS_FIXEDFILEINFO 结构体
type VS_FIXEDFILEINFO struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsMask    uint32
	FileFlags        uint32
	FileOS           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}

// initialize 初始化版本信息
func (i *Info) initialize() error {
	// 转换路径为 UTF16
	pathPtr, err := syscall.UTF16PtrFromString(i.FilePath)
	if err != nil {
		return err
	}

	// 获取版本信息大小
	var handle uintptr
	size, _, err := procGetFileVersionInfoSize.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&handle)),
	)
	if size == 0 {
		return fmt.Errorf("GetFileVersionInfoSize failed: %v", err)
	}

	// 分配内存
	verInfo := make([]byte, size)
	ret, _, err := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		size,
		uintptr(unsafe.Pointer(&verInfo[0])),
	)
	if ret == 0 {
		return fmt.Errorf("GetFileVersionInfo failed: %v", err)
	}

	// 获取固定的文件信息
	var fixedFileInfo *VS_FIXEDFILEINFO
	var uLen uint32
	rootPtr, _ := syscall.UTF16PtrFromString("\\")
	ret, _, err = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&verInfo[0])),
		uintptr(unsafe.Pointer(rootPtr)),
		uintptr(unsafe.Pointer(&fixedFileInfo)),
		uintptr(unsafe.Pointer(&uLen)),
	)
	if ret == 0 {
		return fmt.Errorf("VerQueryValue failed: %v", err)
	}

	// 解析文件版本
	i.FullVersion = fmt.Sprintf("%d.%d.%d.%d",
		(fixedFileInfo.FileVersionMS>>16)&0xffff,
		(fixedFileInfo.FileVersionMS>>0)&0xffff,
		(fixedFileInfo.FileVersionLS>>16)&0xffff,
		(fixedFileInfo.FileVersionLS>>0)&0xffff,
	)
	i.Version = int((fixedFileInfo.FileVersionMS >> 16) & 0xffff)

	i.ProductVersion = fmt.Sprintf("%d.%d.%d.%d",
		(fixedFileInfo.ProductVersionMS>>16)&0xffff,
		(fixedFileInfo.ProductVersionMS>>0)&0xffff,
		(fixedFileInfo.ProductVersionLS>>16)&0xffff,
		(fixedFileInfo.ProductVersionLS>>0)&0xffff,
	)

	// 获取翻译信息
	type langAndCodePage struct {
		language uint16
		codePage uint16
	}

	var lpTranslate *langAndCodePage
	var cbTranslate uint32
	transPtr, _ := syscall.UTF16PtrFromString("\\VarFileInfo\\Translation")
	ret, _, _ = procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&verInfo[0])),
		uintptr(unsafe.Pointer(transPtr)),
		uintptr(unsafe.Pointer(&lpTranslate)),
		uintptr(unsafe.Pointer(&cbTranslate)),
	)

	if ret != 0 && cbTranslate > 0 {
		// 获取所有需要的字符串信息
		stringInfos := map[string]*string{
			"CompanyName":     &i.CompanyName,
			"FileDescription": &i.FileDescription,
			"FileVersion":     &i.FullVersion,
			"LegalCopyright":  &i.LegalCopyright,
			"ProductName":     &i.ProductName,
			"ProductVersion":  &i.ProductVersion,
		}

		for name, ptr := range stringInfos {
			subBlock := fmt.Sprintf("\\StringFileInfo\\%04x%04x\\%s",
				lpTranslate.language, lpTranslate.codePage, name)

			subBlockPtr, _ := syscall.UTF16PtrFromString(subBlock)
			var buffer *uint16
			var bufLen uint32

			ret, _, _ = procVerQueryValue.Call(
				uintptr(unsafe.Pointer(&verInfo[0])),
				uintptr(unsafe.Pointer(subBlockPtr)),
				uintptr(unsafe.Pointer(&buffer)),
				uintptr(unsafe.Pointer(&bufLen)),
			)

			if ret != 0 && bufLen > 0 {
				*ptr = syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(buffer))[:bufLen:bufLen])
			}
		}
	}

	return nil
}
