package memprocfs

// #include <stdlib.h>
// #include "leechcore.h"
// #include "vmmdll.h"
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

type VadEntry struct {
	VaStart       uintptr
	VaEnd         uintptr
	VadType       uint32
	Protection    uint32
	IsImage       bool
	IsFile        bool
	IsPageFile    bool
	IsPrivate     bool
	IsTeb         bool
	IsStack       bool
	IsHeap        bool
	HeapNum       uint32
	CommitCharge  uint32
	MemCommit     bool
	Text          string
	VaFileObject  uintptr
	CVadExPages   uint32
}

func (h *MemProcFS) GetVadMap(pid int32, identifyModules bool) ([]VadEntry, error) {
	fIdentify := C.BOOL(0)
	if identifyModules {
		fIdentify = C.BOOL(1)
	}

	var pVadMap C.PVMMDLL_MAP_VAD
	ok := C.VMMDLL_Map_GetVadU(h.vmDllHandle, C.DWORD(pid), fIdentify, &pVadMap)
	if ok == 0 {
		return nil, fmt.Errorf("failed to get VAD map for pid: %d", pid)
	}
	defer C.VMMDLL_MemFree(C.PVOID(unsafe.Pointer(pVadMap)))

	count := int(C.VadMap_GetCount(pVadMap))
	result := make([]VadEntry, count)
	for i := 0; i < count; i++ {
		e := C.VadMap_GetEntry(pVadMap, C.DWORD(i))
		entry := VadEntry{
			VaStart:      uintptr(e.vaStart),
			VaEnd:        uintptr(e.vaEnd),
			VadType:      uint32(C.VadEntry_GetVadType(e)),
			Protection:   uint32(C.VadEntry_GetProtection(e)),
			IsImage:      C.VadEntry_GetfImage(e) != 0,
			IsFile:       C.VadEntry_GetfFile(e) != 0,
			IsPageFile:   C.VadEntry_GetfPageFile(e) != 0,
			IsPrivate:    C.VadEntry_GetfPrivateMemory(e) != 0,
			IsTeb:        C.VadEntry_GetfTeb(e) != 0,
			IsStack:      C.VadEntry_GetfStack(e) != 0,
			IsHeap:       C.VadEntry_GetfHeap(e) != 0,
			HeapNum:      uint32(C.VadEntry_GetHeapNum(e)),
			CommitCharge: uint32(C.VadEntry_GetCommitCharge(e)),
			MemCommit:    C.VadEntry_GetMemCommit(e) != 0,
			VaFileObject: uintptr(e.vaFileObject),
			CVadExPages:  uint32(e.cVadExPages),
		}
		txt := C.VadEntry_GetText(e)
		if txt != nil {
			entry.Text = C.GoString(txt)
		}
		result[i] = entry
	}
	return result, nil
}

type HeapType uint32

const (
	HeapTypeNA  HeapType = 0 // Unknown
	HeapTypeNT  HeapType = 1 // NT Heap
	HeapTypeSeg HeapType = 2 // Segment Heap
)

type HeapSegmentType uint32

const (
	HeapSegmentNA         HeapSegmentType = 0
	HeapSegmentNTSegment  HeapSegmentType = 1
	HeapSegmentNTLFH      HeapSegmentType = 2
	HeapSegmentNTLarge    HeapSegmentType = 3
	HeapSegmentNTNA       HeapSegmentType = 4
	HeapSegmentSegHeap    HeapSegmentType = 5
	HeapSegmentSegSegment HeapSegmentType = 6
	HeapSegmentSegLarge   HeapSegmentType = 7
	HeapSegmentSegNA      HeapSegmentType = 8
)

type HeapEntry struct {
	Va        uintptr
	Type      HeapType
	Is32      bool
	IHeap     uint32
	HeapNum   uint32
}

type HeapSegmentEntry struct {
	Va    uintptr
	Size  uint32
	Type  HeapSegmentType
	IHeap uint32
}

type HeapMap struct {
	Heaps    []HeapEntry
	Segments []HeapSegmentEntry
}

func (h *MemProcFS) GetHeapMap(pid int32) (*HeapMap, error) {
	var pHeapMap C.PVMMDLL_MAP_HEAP
	ok := C.VMMDLL_Map_GetHeap(h.vmDllHandle, C.DWORD(pid), &pHeapMap)
	if ok == 0 {
		return nil, fmt.Errorf("failed to get heap map for pid: %d", pid)
	}
	defer C.VMMDLL_MemFree(C.PVOID(unsafe.Pointer(pHeapMap)))

	heapCount := int(C.HeapMap_GetCount(pHeapMap))
	heaps := make([]HeapEntry, heapCount)
	for i := 0; i < heapCount; i++ {
		e := C.HeapMap_GetEntry(pHeapMap, C.DWORD(i))
		heaps[i] = HeapEntry{
			Va:      uintptr(e.va),
			Type:    HeapType(e.tp),
			Is32:    e.f32 != 0,
			IHeap:   uint32(e.iHeap),
			HeapNum: uint32(e.dwHeapNum),
		}
	}

	segCount := int(C.HeapMap_GetSegmentCount(pHeapMap))
	segments := make([]HeapSegmentEntry, segCount)
	for i := 0; i < segCount; i++ {
		s := C.HeapMap_GetSegment(pHeapMap, C.DWORD(i))
		segments[i] = HeapSegmentEntry{
			Va:    uintptr(s.va),
			Size:  uint32(s.cb),
			Type:  HeapSegmentType(C.HeapSegment_GetTp(s)),
			IHeap: uint32(C.HeapSegment_GetIHeap(s)),
		}
	}

	return &HeapMap{
		Heaps:    heaps,
		Segments: segments,
	}, nil
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

type VmmDllModuleEntry struct {
	VaBase      uintptr
	VaEntry     uintptr
	CbImageSize uint32
	FWoW64      bool
}

func (h *MemProcFS) GetModuleInfo(pid int32, moduleName string) (*VmmDllModuleEntry, error) {
	cmoduleName := C.CString(moduleName)
	defer C.free(unsafe.Pointer(cmoduleName))

	var pModuleEntry *C.VMMDLL_MAP_MODULEENTRY
	ppModuleEntry := (*C.PVMMDLL_MAP_MODULEENTRY)(unsafe.Pointer(&pModuleEntry))

	base := C.VMMDLL_Map_GetModuleFromNameU(
		h.vmDllHandle,
		C.DWORD(pid),
		cmoduleName,
		ppModuleEntry,
		0,
	)
	if base == 0 {
		return nil, fmt.Errorf("failed to get module base: %s", moduleName)
	}
	return &VmmDllModuleEntry{
		VaBase:      uintptr(pModuleEntry.vaBase),
		VaEntry:     uintptr(pModuleEntry.vaEntry),
		CbImageSize: uint32(pModuleEntry.cbImageSize),
		FWoW64:      pModuleEntry.fWoW64 == 1,
	}, nil
}
