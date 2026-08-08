package system

import (
	"ps1emu/cpu"
	"ps1emu/memory"
)

type System struct {
	Bus *memory.Bus
	CPU *cpu.CPU
}

func New() *System {
	bus := memory.New()
	return &System{
		Bus: bus,
		CPU: cpu.New(bus),
	}
}

func NewWithBIOS(data []byte) (*System, error) {
	bus, err := memory.NewWithBIOS(data)
	if err != nil {
		return nil, err
	}

	return &System{
		Bus: bus,
		CPU: cpu.New(bus),
	}, nil
}

func (s *System) Step() uint32 {
	return s.CPU.Step()
}
