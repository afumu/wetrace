package util

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func Is64Bit(handle windows.Handle) (bool, error) {
	var is32Bit bool
	if err := windows.IsWow64Process(handle, &is32Bit); err != nil {
		return false, fmt.Errorf("检查进程位数失败: %w", err)
	}
	return !is32Bit, nil
}
