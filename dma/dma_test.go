package dma_test

import (
	"encoding/binary"
	"testing"

	"ps1emu/dma"
)

type MockBus struct {
	ram []byte
}

func NewMockBus(size int) *MockBus {
	return &MockBus{ram: make([]byte, size)}
}

func (m *MockBus) Read32(addr uint32) uint32 {
	addr &= 0x00FFFFFF
	if int(addr+4) > len(m.ram) {
		return 0
	}
	return binary.LittleEndian.Uint32(m.ram[addr : addr+4])
}

func (m *MockBus) Write32(addr uint32, val uint32) {
	addr &= 0x00FFFFFF
	if int(addr+4) <= len(m.ram) {
		binary.LittleEndian.PutUint32(m.ram[addr:addr+4], val)
	}
}

func TestDMANew(t *testing.T) {
	ctrl := dma.New()

	if ctrl.DPCR != 0x07654321 {
		t.Errorf("Expected initial DPCR=0x07654321, got 0x%08X", ctrl.DPCR)
	}

	// Read DPCR register offset 0x70
	if val := ctrl.Read32(0x70); val != 0x07654321 {
		t.Errorf("Read32(0x70) expected 0x07654321, got 0x%08X", val)
	}
}

func TestDMAChannelRegisters(t *testing.T) {
	ctrl := dma.New()
	bus := NewMockBus(1024 * 1024)

	// Channel 6 (OTC) MADR offset: 0x60
	ctrl.Write32(0x60, 0x00100000, bus)
	if val := ctrl.Read32(0x60); val != 0x00100000 {
		t.Errorf("OTC MADR expected 0x00100000, got 0x%08X", val)
	}

	// Channel 6 BCR offset: 0x64
	ctrl.Write32(0x64, 0x0004, bus)
	if val := ctrl.Read32(0x64); val != 0x0004 {
		t.Errorf("OTC BCR expected 0x0004, got 0x%08X", val)
	}
}

func TestDMAOTCTransfer(t *testing.T) {
	ctrl := dma.New()
	bus := NewMockBus(1024 * 1024)

	// Set MADR = 0x1000 (4096)
	ctrl.Write32(0x60, 0x00001000, bus)
	// Set BCR = 4 words
	ctrl.Write32(0x64, 4, bus)

	// Set CHCR (Sync 0, Step decrement -4, Direction ToRAM, Trigger bit 28, Enable bit 24)
	// Bit 0 = 0 (ToRAM), Bit 1 = 1 (-4 step), Bit 9..8 = 00 (Sync 0), Bit 24 = 1, Bit 28 = 1
	chcrVal := uint32((1 << 28) | (1 << 24) | (1 << 1))
	ctrl.Write32(0x68, chcrVal, bus)

	// Verify OTC linked table written to RAM
	// Entry 0 @ 0x1000 -> points to 0x0DFC
	// Entry 1 @ 0x0FC -> points to 0x0FF8
	// Entry 2 @ 0x0F8 -> points to 0x0FF4
	// Entry 3 @ 0x0F4 -> 0x00FFFFFF (End marker)
	word0 := bus.Read32(0x1000)
	word3 := bus.Read32(0x0FF4)

	if word0 != 0x00000FFC {
		t.Errorf("OTC Word 0 expected 0x00000FFC, got 0x%08X", word0)
	}
	if word3 != 0x00FFFFFF {
		t.Errorf("OTC Word 3 (End marker) expected 0x00FFFFFF, got 0x%08X", word3)
	}
}

func TestDMAInterrupts(t *testing.T) {
	ctrl := dma.New()
	bus := NewMockBus(1024 * 1024)

	// Enable Channel 6 interrupt (Bit 22 in DICR) & Master Enable (Bit 23)
	ctrl.Write32(0x74, (1<<22)|(1<<23), bus)

	// Perform OTC Transfer on Channel 6
	ctrl.Write32(0x60, 0x1000, bus)
	ctrl.Write32(0x64, 2, bus)
	ctrl.Write32(0x68, (1<<28)|(1<<24)|(1<<1), bus)

	// Verify IRQ master flag (Bit 31) and Channel 6 flag (Bit 30) are set
	dicr := ctrl.Read32(0x74)
	if (dicr & (1 << 31)) == 0 {
		t.Errorf("Master IRQ bit 31 expected to be set, DICR=0x%08X", dicr)
	}
	if (dicr & (1 << 30)) == 0 {
		t.Errorf("Channel 6 IRQ flag bit 30 expected to be set, DICR=0x%08X", dicr)
	}

	// Clear Channel 6 flag by writing 1 to bit 30
	ctrl.Write32(0x74, (1<<30)|(1<<22)|(1<<23), bus)

	dicrAfter := ctrl.Read32(0x74)
	if (dicrAfter & (1 << 30)) != 0 {
		t.Errorf("Channel 6 IRQ flag should be cleared, DICR=0x%08X", dicrAfter)
	}
}
