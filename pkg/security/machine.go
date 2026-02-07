package security

import (
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"sort"
	"strings"
)

// GetMachineID generates a unique identifier for the machine
// Priority: UUID > CPU+Disk(Sorted) > Hostname
func GetMachineID() (string, error) {
	// 1. 优先尝试获取主板 UUID (最稳定，不受硬件插拔顺序影响)
	uuid, err := getPlatformUUID()
	if err == nil && uuid != "" {
		hash := sha256.Sum256([]byte(uuid))
		return hex.EncodeToString(hash[:]), nil
	}

	// 2. 回退方案：CPU + 硬盘序列号
	cpuID, err := getCPUSerial()
	if err != nil {
		// Log error if needed, but continue
	}

	diskID, err := getDiskSerial()
	if err != nil {
		// Log error if needed
	}

	// 3. 最后的备选：Hostname
	if cpuID == "" && diskID == "" {
		cmd := exec.Command("hostname")
		out, err := cmd.Output()
		if err == nil {
			return hex.EncodeToString(sha256.New().Sum(out)), nil
		}
		return "", err // Return last error if everything failed
	}

	// Combine and Hash
	data := cpuID + "|" + diskID
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

func getPlatformUUID() (string, error) {
	// 获取主板 UUID
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-WmiObject -Class Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getCPUSerial() (string, error) {
	// 使用 PowerShell 获取 ProcessorId
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-WmiObject -Class Win32_Processor | Select-Object -ExpandProperty ProcessorId")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getDiskSerial() (string, error) {
	// 使用 PowerShell 获取 Disk SerialNumber
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-WmiObject -Class Win32_DiskDrive | Select-Object -ExpandProperty SerialNumber")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 处理多硬盘情况：分割、去空、排序
	raw := string(out)
	lines := strings.Split(raw, "\n")
	var serials []string
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s != "" {
			serials = append(serials, s)
		}
	}

	// 关键步骤：排序。保证无论 WMI 以何种顺序返回硬盘，生成的 ID 都是一致的。
	sort.Strings(serials)

	return strings.Join(serials, ","), nil
}
