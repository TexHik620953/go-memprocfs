package memprocfs

import (
	"unsafe"
)

type SteppedScatterReadTask struct {
	scatter *ScatterReadTask

	steps int

	stepsGrid [][]uintptr
	finalGrid []unsafe.Pointer
	finalSize int32
}

func (h *MemProcFS) NewSteppedScatterTask(pid int32, readsPerStep int, steps int, finalSize int32) (*SteppedScatterReadTask, error) {
	scatter, err := h.NewScatterTask(pid)
	if err != nil {
		return nil, err
	}
	t := &SteppedScatterReadTask{
		scatter:   scatter,
		steps:     steps,
		stepsGrid: make([][]uintptr, readsPerStep),
		finalGrid: make([]unsafe.Pointer, readsPerStep),
		finalSize: finalSize,
	}
	return t, nil
}
func (h *SteppedScatterReadTask) Close() {
	h.scatter.Close()
}
func (h *SteppedScatterReadTask) Set(id int, addr []uintptr, target unsafe.Pointer) {
	h.stepsGrid[id] = addr
	h.finalGrid[id] = target
}

func (h *SteppedScatterReadTask) Execute() error {
	tempAddresses := make([]uintptr, len(h.stepsGrid))
	for i := 0; i < len(h.stepsGrid); i++ {
		tempAddresses[i] = h.stepsGrid[i][0]
	}
	tempResults := make([]uintptr, len(h.stepsGrid))

	for stepId := 1; stepId < h.steps; stepId++ {
		for taskId := 0; taskId < len(tempAddresses); taskId++ {
			h.scatter.AddRead(tempAddresses[taskId], 8, unsafe.Pointer(&tempResults[taskId]))
		}
		err := h.scatter.Execute()
		if err != nil {
			return err
		}
		for taskId := 0; taskId < len(tempAddresses); taskId++ {
			tempAddresses[taskId] = tempResults[taskId] + h.stepsGrid[taskId][stepId]
		}
	}

	//Final read
	for taskId := 0; taskId < len(h.finalGrid); taskId++ {
		h.scatter.AddRead(tempAddresses[taskId], h.finalSize, h.finalGrid[taskId])
	}
	err := h.scatter.Execute()
	if err != nil {
		return err
	}
	return nil
}
