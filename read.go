package memprocfs

// #include <stdlib.h>
// #include "leechcore.h"
// #include "vmmdll.h"
import "C"
import (
	"encoding/binary"
	"math"
)

func (h *MemProcFS) MemReadChain(pid int32, addr []uintptr, size int32) ([]byte, error) {
	// Resolve chain target addr
	tempAddr := addr[0]
	for i := 1; i < len(addr); i++ {
		buff, err := h.MemRead(pid, tempAddr, 8)
		if err != nil {
			return nil, err
		}
		tempAddr = uintptr(binary.LittleEndian.Uint64(buff))
		tempAddr = tempAddr + addr[i]
	}

	return h.MemRead(pid, tempAddr, size)
}

func (h *MemProcFS) ReadInt32(pid int32, addr []uintptr) (int32, error) {
	data, err := h.MemReadChain(pid, addr, 4)
	if err != nil {
		return 0, err
	}

	var value int32
	value |= int32(data[0])
	value |= int32(data[1]) << 8
	value |= int32(data[2]) << 16
	value |= int32(data[3]) << 24
	return value, nil
}
func (h *MemProcFS) ReadInt16(pid int32, addr []uintptr) (int16, error) {
	data, err := h.MemReadChain(pid, addr, 2)
	if err != nil {
		return 0, err
	}

	var value int16
	value |= int16(data[0])
	value |= int16(data[1]) << 8
	return value, nil
}

func (h *MemProcFS) ReadInt64(pid int32, addr []uintptr) (int64, error) {
	data, err := h.MemReadChain(pid, addr, 8)
	if err != nil {
		return 0, err
	}

	var value int64
	value |= int64(data[0])
	value |= int64(data[1]) << 8
	value |= int64(data[2]) << 16
	value |= int64(data[3]) << 24
	value |= int64(data[3]) << 32
	value |= int64(data[3]) << 40
	value |= int64(data[3]) << 48
	value |= int64(data[3]) << 56
	return value, nil
}

func (h *MemProcFS) ReadUInt32(pid int32, addr []uintptr) (uint32, error) {
	data, err := h.MemReadChain(pid, addr, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}
func (h *MemProcFS) ReadUInt64(pid int32, addr []uintptr) (uint64, error) {
	data, err := h.MemReadChain(pid, addr, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(data), nil
}
func (h *MemProcFS) ReadPtr(pid int32, addr []uintptr) (uintptr, error) {
	data, err := h.MemReadChain(pid, addr, 8)
	if err != nil {
		return 0, err
	}
	return uintptr(binary.LittleEndian.Uint64(data)), nil
}

func (h *MemProcFS) ReadFloat32(pid int32, addr []uintptr) (float32, error) {
	temp, err := h.ReadUInt32(pid, addr)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(temp), nil
}
func (h *MemProcFS) ReadFloat64(pid int32, addr []uintptr) (float64, error) {
	temp, err := h.ReadUInt64(pid, addr)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(temp), nil
}

func (h *MemProcFS) ReadBool(pid int32, addr []uintptr) (bool, error) {
	data, err := h.MemReadChain(pid, addr, 1)
	if err != nil {
		return false, err
	}
	return data[0] > 0, nil
}
