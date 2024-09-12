package memprocfs

// #include <stdlib.h>
// #include "leechcore.h"
// #include "vmmdll.h"
import "C"
import (
	"bytes"
	"fmt"
	"reflect"
	"unsafe"
)

type MemProcFS struct {
	vmDllHandle C.VMM_HANDLE
}

// Initialize VMMDLL
// for ex New("-device", "fpga", "-memmap", "memmap.txt")
func New(args ...string) (*MemProcFS, error) {
	if len(args) == 0 {
		//Default
		args = []string{"-device", "fpga", "-memmap", "memmap.txt"}
	}
	cArgs := make([]C.LPCSTR, len(args))
	for i, arg := range args {
		cArgs[i] = C.LPCSTR(C.CString(arg))
		defer C.free(unsafe.Pointer(cArgs[i]))
	}

	vmmHandle := C.VMMDLL_Initialize(C.DWORD(len(args)), &cArgs[0])
	if vmmHandle == nil {
		return nil, fmt.Errorf("failed to initialize FPGA")
	}

	return &MemProcFS{
		vmDllHandle: vmmHandle,
	}, nil
}
func (h *MemProcFS) Close() {
	C.VMMDLL_Close(h.vmDllHandle)
}

func (h *MemProcFS) GetPidByName(name string) (int32, error) {
	pid := C.DWORD(0)
	ok := C.VMMDLL_PidGetFromName(h.vmDllHandle, C.LPSTR(C.CString(name)), &pid)

	if pid == 0 || ok == 0 {
		return 0, fmt.Errorf("failed to find process by name: %s", name)
	}
	return int32(pid), nil
}

type VMMDLLProcessInformation struct {
	Magic         uint64
	WVersion      uint16
	WSize         uint16
	TpMemoryModel uint32
	TpSystem      uint32
	FUserOnly     bool
	DwPID         uint32
	DwPPID        uint32
	DwState       uint32
	SzName        string
	SzNameLong    string
	PaDTB         uint64
	PaDTB_UserOpt uint64
	Win           struct {
		VaEPROCESS     uint64
		VaPEB          uint64
		Reserved1      uint64
		FWow64         bool
		VaPEB32        uint32
		DwSessionId    uint32
		QwLUID         uint64
		SzSID          [260]byte
		IntegrityLevel uint32
	}
}

func (h *MemProcFS) GetProcessInfo(pid int32) (*VMMDLLProcessInformation, error) {
	size := C.SIZE_T(0)
	C.VMMDLL_ProcessGetInformation(h.vmDllHandle, C.DWORD(pid), nil, &size)

	pInfo := C.VMMDLL_PROCESS_INFORMATION{
		magic:    C.VMMDLL_PROCESS_INFORMATION_MAGIC,
		wVersion: C.VMMDLL_PROCESS_INFORMATION_VERSION,
	}
	ok := C.VMMDLL_ProcessGetInformation(h.vmDllHandle, C.DWORD(pid), &pInfo, &size)
	if ok == 0 {
		return nil, fmt.Errorf("failed to get process info: %d", pid)
	}
	result := &VMMDLLProcessInformation{
		Magic:         uint64(pInfo.magic),
		WVersion:      uint16(pInfo.wVersion),
		WSize:         uint16(pInfo.wSize),
		TpMemoryModel: uint32(pInfo.tpMemoryModel),
		TpSystem:      uint32(pInfo.tpSystem),
		FUserOnly:     pInfo.fUserOnly != 0,
		DwPID:         uint32(pInfo.dwPID),
		DwPPID:        uint32(pInfo.dwPPID),
		DwState:       uint32(pInfo.dwState),
		PaDTB:         uint64(pInfo.paDTB),
		PaDTB_UserOpt: uint64(pInfo.paDTB_UserOpt),
		Win: struct {
			VaEPROCESS     uint64
			VaPEB          uint64
			Reserved1      uint64
			FWow64         bool
			VaPEB32        uint32
			DwSessionId    uint32
			QwLUID         uint64
			SzSID          [260]byte
			IntegrityLevel uint32
		}{
			VaEPROCESS:     uint64(pInfo.win.vaEPROCESS),
			VaPEB:          uint64(pInfo.win.vaPEB),
			Reserved1:      uint64(pInfo.win._Reserved1),
			FWow64:         pInfo.win.fWow64 != 0,
			VaPEB32:        uint32(pInfo.win.vaPEB32),
			DwSessionId:    uint32(pInfo.win.dwSessionId),
			QwLUID:         uint64(pInfo.win.qwLUID),
			IntegrityLevel: uint32(pInfo.win.IntegrityLevel),
			SzSID:          *(*[260]byte)(unsafe.Pointer(&pInfo.win.szSID)),
		},
	}

	szName := (*(*[16]byte)(unsafe.Pointer(&pInfo.szName)))[:]
	n := bytes.Index(szName, []byte{0})
	result.SzName = string(szName[:n])

	szLongName := (*(*[64]byte)(unsafe.Pointer(&pInfo.szNameLong)))[:]
	n = bytes.Index(szLongName, []byte{0})
	result.SzNameLong = string(szLongName[:n])
	return result, nil
}

func (h *MemProcFS) MemWrite(pid int32, addr uintptr, data []byte) error {
	// TODO: Provide data array by pointer into function to avoid copying
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ok := C.VMMDLL_MemWrite(
		h.vmDllHandle,
		C.DWORD(pid),
		C.ULONG64(addr),
		C.PBYTE(unsafe.Pointer(sh.Data)),
		C.DWORD(len(data)))

	if ok == 0 {
		return fmt.Errorf("failed to MemWrite: %d %d", pid, addr)
	}
	return nil
}

func (h *MemProcFS) MemRead(pid int32, addr uintptr, size int32) ([]byte, error) {

	if !IsValidAddress(addr) {
		return nil, fmt.Errorf("failed to read, invalid address: %X", addr)
	}
	buff := make([]byte, size, size)
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buff))
	ok := C.VMMDLL_MemReadEx(
		h.vmDllHandle,
		C.DWORD(pid),
		C.ULONG64(addr),
		C.PBYTE(unsafe.Pointer(sh.Data)),
		C.DWORD(size),
		nil,
		C.ULONG64(0x0003))

	if ok == 0 {
		return nil, fmt.Errorf("failed to MemRead: %d %X", pid, addr)
	}

	return buff, nil
}

func (h *MemProcFS) MemReadToBuff(pid int32, addr uintptr, size int32, ptr unsafe.Pointer) error {

	if !IsValidAddress(addr) {
		return fmt.Errorf("failed to read, invalid address: %X", addr)
	}
	ok := C.VMMDLL_MemReadEx(
		h.vmDllHandle,
		C.DWORD(pid),
		C.ULONG64(addr),
		C.PBYTE(ptr),
		C.DWORD(size),
		nil,
		C.ULONG64(0x0003))

	if ok == 0 {
		return fmt.Errorf("failed to MemRead: %d %X", pid, addr)
	}
	return nil
}
