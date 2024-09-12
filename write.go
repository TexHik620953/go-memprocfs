package memprocfs

import (
	"encoding/binary"
	"math"
)

func (h *MemProcFS) WriteFloat32(pid int32, addr uintptr, value float32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(value))

	return h.MemWrite(pid, addr, buf[:])
}

func (h *MemProcFS) WriteBool(pid int32, addr uintptr, value bool) error {
	val := 0
	if value {
		val = 1
	}
	buf := []byte{byte(val)}

	return h.MemWrite(pid, addr, buf[:])
}

func (h *MemProcFS) WriteInt32(pid int32, addr uintptr, value int32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(value))

	return h.MemWrite(pid, addr, buf[:])
}
