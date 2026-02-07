package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procFindWindowW              = user32.NewProc("FindWindowW")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procEnumWindows              = user32.NewProc("EnumWindows")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
)

const (
	CREATE_NEW_CONSOLE       = 0x00000010
	CREATE_NEW_PROCESS_GROUP = 0x00000200
)

// Helper function to check if a file/directory exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper to read a string value from registry
func readRegistryString(rootKey registry.Key, keyPath string, valueName string) (string, error) {
	k, err := registry.OpenKey(rootKey, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	if valueName == "" { // Default value
		val, _, err := k.GetStringValue("")
		return val, err
	}
	val, _, err := k.GetStringValue(valueName)
	return val, err
}

type ProcessManager struct {
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

func (pm *ProcessManager) IsProcessRunning(name string) bool {
	pid, _ := pm.GetProcessId(name)
	return pid != 0
}

func (pm *ProcessManager) GetProcessId(name string) (uint32, error) {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer syscall.CloseHandle(snapshot)

	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	if err = syscall.Process32First(snapshot, &procEntry); err != nil {
		return 0, err
	}

	for {
		exeName := syscall.UTF16ToString(procEntry.ExeFile[:])
		if strings.EqualFold(exeName, name) {
			return procEntry.ProcessID, nil
		}
		if err = syscall.Process32Next(snapshot, &procEntry); err != nil {
			break
		}
	}
	return 0, fmt.Errorf("process not found")
}

func (pm *ProcessManager) KillProcess(name string) error {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(snapshot)

	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	if err = syscall.Process32First(snapshot, &procEntry); err != nil {
		return err
	}

	for {
		exeName := syscall.UTF16ToString(procEntry.ExeFile[:])
		if strings.EqualFold(exeName, name) {
			hProcess, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, procEntry.ProcessID)
			if err == nil {
				syscall.TerminateProcess(hProcess, 0)
				syscall.CloseHandle(hProcess)
			}
		}
		if err = syscall.Process32Next(snapshot, &procEntry); err != nil {
			break
		}
	}
	return nil
}

func (pm *ProcessManager) LaunchWeChat(path string) error {
	// Use os/exec to start detached process
	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NEW_CONSOLE | CREATE_NEW_PROCESS_GROUP,
	}
	return cmd.Start()
}

func (pm *ProcessManager) FindWeChatPath() string {
	// 1. Check Registry
	path := pm.findWeChatFromRegistry()
	if path != "" && fileExists(path) {
		return path
	}

	// 2. Check standard paths
	path = pm.findWeChatFromCommonPaths()
	if path != "" && fileExists(path) {
		return path
	}

	return ""
}

func (pm *ProcessManager) findWeChatFromRegistry() string {
	// 1. Uninstall
	path := pm.findWeChatFromUninstall()
	if path != "" {
		return path
	}
	// 2. App Paths
	path = pm.findWeChatFromAppPaths()
	if path != "" {
		return path
	}
	// 3. Tencent Registry
	path = pm.findWeChatFromTencentRegistry()
	if path != "" {
		return path
	}
	return ""
}

func (pm *ProcessManager) findWeChatFromUninstall() string {
	uninstallKeys := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\WeChat`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\WeChat`,
	}
	for _, key := range uninstallKeys {
		path := pm.readInstallLocationFromKey(registry.LOCAL_MACHINE, key)
		if path != "" {
			return path
		}
	}
	return ""
}

func (pm *ProcessManager) findWeChatFromAppPaths() string {
	appNames := []string{"WeChat.exe", "Weixin.exe"}
	rootKeys := []registry.Key{registry.LOCAL_MACHINE, registry.CURRENT_USER}

	for _, rootKey := range rootKeys {
		for _, appName := range appNames {
			keyPath := `SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths\` + appName
			path, err := readRegistryString(rootKey, keyPath, "")
			if err == nil && path != "" && fileExists(path) {
				return path
			}
		}
	}
	return ""
}

func (pm *ProcessManager) findWeChatFromTencentRegistry() string {
	keyPaths := []string{
		`Software\Tencent\WeChat`,
		`Software\Tencent\bugReport\WeChatWindows`,
		`Software\WOW6432Node\Tencent\WeChat`,
		`Software\Tencent\Weixin`,
	}
	valueNames := []string{"InstallPath", "Install", "Path", "InstallDir"}
	rootKeys := []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE}

	for _, keyPath := range keyPaths {
		for _, rootKey := range rootKeys {
			for _, valueName := range valueNames {
				// Try reading value
				result, err := readRegistryString(rootKey, keyPath, valueName)
				if err == nil && result != "" {
					// Check if result ends with .exe, if not try to append
					lowerResult := strings.ToLower(result)
					if !strings.HasSuffix(lowerResult, ".exe") {
						if p := pm.checkDirForExe(result); p != "" {
							return p
						}
					}
					if fileExists(result) {
						return result
					}
				}
			}
		}
	}
	return ""
}

func (pm *ProcessManager) readInstallLocationFromKey(rootKey registry.Key, subKey string) string {
	valueNames := []string{
		"InstallLocation",
		"InstallPath",
		"DisplayIcon",
		"UninstallString",
		"InstallDir",
	}

	for _, valueName := range valueNames {
		result, err := readRegistryString(rootKey, subKey, valueName)
		if err != nil || result == "" {
			continue
		}

		exePath := result
		if valueName == "UninstallString" || valueName == "DisplayIcon" {
			parts := strings.Split(exePath, ",")
			if len(parts) > 0 {
				exePath = strings.TrimSpace(parts[0])
				exePath = strings.Trim(exePath, "\"")
			}
		}

		lowerPath := strings.ToLower(exePath)
		if strings.HasSuffix(lowerPath, ".exe") {
			if fileExists(exePath) {
				return exePath
			}
			dir := filepath.Dir(exePath)
			if p := pm.checkDirForExe(dir); p != "" {
				return p
			}
		} else {
			if p := pm.checkDirForExe(exePath); p != "" {
				return p
			}
		}
	}
	return ""
}

func (pm *ProcessManager) checkDirForExe(dir string) string {
	weixin := filepath.Join(dir, "Weixin.exe")
	if fileExists(weixin) {
		return weixin
	}
	wechat := filepath.Join(dir, "WeChat.exe")
	if fileExists(wechat) {
		return wechat
	}
	return ""
}

func (pm *ProcessManager) findWeChatFromCommonPaths() string {
	drives := []string{"C", "D", "E", "F"}
	commonPaths := []string{
		`\Program Files\Tencent\WeChat\WeChat.exe`,
		`\Program Files (x86)\Tencent\WeChat\WeChat.exe`,
		`\Program Files\Tencent\Weixin\Weixin.exe`,
		`\Program Files (x86)\Tencent\Weixin\Weixin.exe`,
	}

	for _, drive := range drives {
		for _, commonPath := range commonPaths {
			fullPath := drive + ":" + commonPath
			if fileExists(fullPath) {
				return fullPath
			}
		}
	}
	return ""
}

func (pm *ProcessManager) WaitForWeChatWindow(timeoutSeconds int) bool {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	for time.Now().Before(deadline) {
		pid := pm.FindMainWeChatPid()
		if pid != 0 {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func (pm *ProcessManager) FindMainWeChatPid() uint32 {
	var foundPid uint32 = 0

	cb := syscall.NewCallback(func(hwnd syscall.Handle, lparam uintptr) uintptr {
		// Optimization: if we already found it, stop (return FALSE = 0)
		// But NewCallback logic might need to handle concurrency or just simple variable capture
		if foundPid != 0 {
			return 0
		}

		// Optional: Filter invisible windows like C++ does in EnumWechatTopWindowProc,
		// but FindMainWeChatPid in C++ actually uses EnumWindowsProc which *doesn't* check visibility strictly?
		// Let's check C++ src/process_manager.cpp:
		// ProcessManager::EnumWindowsProc: doesn't check IsWindowVisible.
		// ProcessManager::EnumWechatTopWindowProc: checks IsWindowVisible.
		//
		// Since we want to find the PID, getting *any* window belonging to WeChat is good enough usually.
		// But let's check title.

		length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
		if length == 0 {
			return 1 // Continue
		}

		buf := make([]uint16, length+1)
		procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
		title := syscall.UTF16ToString(buf)

		if strings.Contains(title, "微信") ||
			strings.Contains(strings.ToLower(title), "wechat") ||
			strings.Contains(strings.ToLower(title), "weixin") {

			var pid uint32
			procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
			if pid != 0 {
				foundPid = pid
				return 0 // Stop enumeration
			}
		}

		return 1 // Continue
	})

	procEnumWindows.Call(cb, 0)

	return foundPid
}

func (pm *ProcessManager) GetWeChatVersion(wechatDir string) string {
	// Reading version resource is complex in Go without extra libraries.
	// We can try to parse the file version info but it requires "version.dll" APIs.
	// For now, we can skip or return "Unknown"
	return ""
}
