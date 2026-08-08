package system

import (
	"testing"

	"ps1emu/cpu"
	"ps1emu/memory"
)

func TestSystemCPUCanExecuteThroughRealBus(t *testing.T) {
	sys := New()
	sys.Bus.Write32(0x00000000, 0x34021234)
	sys.CPU.SetPC(0x00000000)

	result := sys.CPU.StepResult()

	if result.Instruction != 0x34021234 {
		t.Fatalf("instruction = 0x%08X, want 0x34021234", result.Instruction)
	}

	if got := sys.CPU.Reg(2); got != 0x1234 {
		t.Fatalf("R2 = 0x%08X, want 0x00001234", got)
	}
}

func TestSystemBootsFromBIOSResetVector(t *testing.T) {
	bios := make([]byte, memory.BIOSSize)
	bios[0] = 0x34
	bios[1] = 0x12
	bios[2] = 0x02
	bios[3] = 0x34

	sys, err := NewWithBIOS(bios)
	if err != nil {
		t.Fatalf("NewWithBIOS returned error: %v", err)
	}

	result := sys.CPU.StepResult()

	if result.PC != cpu.ResetVector {
		t.Fatalf("result.PC = 0x%08X, want 0x%08X", result.PC, uint32(cpu.ResetVector))
	}

	if got := sys.CPU.Reg(2); got != 0x1234 {
		t.Fatalf("R2 = 0x%08X, want 0x00001234", got)
	}
}
