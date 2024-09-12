package memprocfs

// #include "utils.hpp"
import "C"
import (
	"fmt"
	"unsafe"
)

func IsValidAddress(address uintptr) bool {
	return !(address == 0 || address == 0xcccccccccccccccc || address >= 0x7fffffffffff)
}

func (h *MemProcFS) FixCr3(pid int32, processName string) error {
	cstr := C.CString(processName)
	defer C.free(unsafe.Pointer(cstr))

	ok := C.FixCr3(h.vmDllHandle, C.DWORD(pid), cstr)
	if ok == 0 {
		return fmt.Errorf("failed to fix cr3")
	}
	return nil
}

func (h *MemProcFS) GetModuleBase(pid int32, moduleName string) (uintptr, error) {
	cmoduleName := C.CString(moduleName)
	defer C.free(unsafe.Pointer(cmoduleName))

	base := C.VMMDLL_ProcessGetModuleBaseU(
		h.vmDllHandle,
		C.DWORD(pid),
		cmoduleName,
	)
	if base == 0 {
		return 0, fmt.Errorf("failed to get module base: %s", moduleName)
	}
	return uintptr(base), nil
}
