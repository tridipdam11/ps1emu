package main

import (
	"testing"

	"ps1emu/system"
)

func TestSystemDefaultInit(t *testing.T) {
	sys := system.New()
	if sys == nil {
		t.Fatal("Expected non-nil system instance")
	}
	if sys.CPU == nil {
		t.Fatal("Expected non-nil CPU instance")
	}
	if sys.CPU.PC != 0xBFC00000 {
		t.Errorf("Expected PC to start at 0xBFC00000, got 0x%08X", sys.CPU.PC)
	}

	res := sys.CPU.StepResult()
	if res.Cycles != 1 {
		t.Errorf("Expected 1 cycle for initial step, got %d", res.Cycles)
	}
}
