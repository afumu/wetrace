package process

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modkernel32           = windows.NewLazySystemDLL("kernel32.dll")
	procVirtualQueryEx    = modkernel32.NewProc("VirtualQueryEx")
	procReadProcessMemory = modkernel32.NewProc("ReadProcessMemory")
)

const (
	MEM_COMMIT  = 0x1000
	MEM_PRIVATE = 0x20000
)

// MemoryRegion represents a continuous memory block
type MemoryRegion struct {
	BaseAddress uintptr
	RegionSize  uintptr
}

// MEMORY_BASIC_INFORMATION structure for VirtualQueryEx
type MEMORY_BASIC_INFORMATION struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

// OpenProcess opens an existing local process object.
func OpenProcess(pid uint32) (windows.Handle, error) {
	// PROCESS_QUERY_INFORMATION (0x0400) | PROCESS_VM_READ (0x0010)
	const desiredAccess = 0x0410
	return windows.OpenProcess(desiredAccess, false, pid)
}

// GetMemoryRegions returns a list of committed private memory regions for the process.
func GetMemoryRegions(hProcess windows.Handle) ([]MemoryRegion, error) {
	var regions []MemoryRegion
	var address uintptr = 0
	var mbi MEMORY_BASIC_INFORMATION
	mbiSize := unsafe.Sizeof(mbi)

	for {
		ret, _, _ := procVirtualQueryEx.Call(
			uintptr(hProcess),
			address,
			uintptr(unsafe.Pointer(&mbi)),
			mbiSize,
		)
		if ret == 0 {
			break
		}

		if mbi.State == MEM_COMMIT && mbi.Type == MEM_PRIVATE {
			regions = append(regions, MemoryRegion{
				BaseAddress: mbi.BaseAddress,
				RegionSize:  mbi.RegionSize,
			})
		}

		// Calculate next address
		nextAddress := address + mbi.RegionSize
		if nextAddress <= address {
			// Wrap around or overflow, stop
			break
		}
		address = nextAddress
	}
	return regions, nil
}

// ReadProcessMemory reads data from an area of memory in a specified process.
func ReadProcessMemory(hProcess windows.Handle, address uintptr, size uintptr) ([]byte, error) {
	buf := make([]byte, size)
	var bytesRead uintptr
	ret, _, _ := procReadProcessMemory.Call(
		uintptr(hProcess),
		address,
		uintptr(unsafe.Pointer(&buf[0])),
		size,
		uintptr(unsafe.Pointer(&bytesRead)),
	)
	if ret == 0 {
		return nil, windows.GetLastError()
	}
	// Return only the bytes actually read
	return buf[:bytesRead], nil
}
