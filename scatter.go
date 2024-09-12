package memprocfs

// #include <stdlib.h>
// #include "leechcore.h"
// #include "vmmdll.h"
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

const scatter_flags = 0x0001 | 0x0002

type ScatterReadTask struct {
	mfs           *MemProcFS
	scatterHandle C.VMMDLL_SCATTER_HANDLE
	pid           int32
	locker        sync.Mutex
	requests      int
}

func (h *MemProcFS) NewScatterTask(pid int32) (*ScatterReadTask, error) {
	task := &ScatterReadTask{
		pid: pid,
		mfs: h,
	}

	task.scatterHandle = C.VMMDLL_Scatter_Initialize(h.vmDllHandle, C.DWORD(pid), scatter_flags)

	if task.scatterHandle == nil {
		return nil, fmt.Errorf("failed to create scatter task")
	}
	return task, nil
}

func (h *ScatterReadTask) Close() {
	h.locker.Lock()
	defer h.locker.Unlock()
	C.VMMDLL_Scatter_CloseHandle(h.scatterHandle)
}
func (h *ScatterReadTask) reset() error {
	h.locker.Lock()
	defer h.locker.Unlock()
	h.requests = 0
	ok := C.VMMDLL_Scatter_Clear(h.scatterHandle, C.DWORD(h.pid), scatter_flags)
	if ok == 0 {
		return fmt.Errorf("failed to clear scatter task")
	}
	return nil
}

func (h *ScatterReadTask) AddRead(addr uintptr, size int32, target unsafe.Pointer) {
	h.locker.Lock()
	h.requests++
	C.VMMDLL_Scatter_PrepareEx(h.scatterHandle, C.QWORD(addr), C.DWORD(size), C.PBYTE(target), nil)
	h.locker.Unlock()
	if h.requests > 300 {
		h.Execute()
	}
}

func (h *ScatterReadTask) Execute() error {
	if h.requests == 0 {
		return nil
	}
	defer h.reset()
	h.locker.Lock()
	ok := C.VMMDLL_Scatter_ExecuteRead(h.scatterHandle)
	h.locker.Unlock()
	if ok == 0 {
		return fmt.Errorf("failed to execute scatter read")
	}
	return nil
}
