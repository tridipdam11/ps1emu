package memory

import "testing"

func TestRAMMirrorsAcrossVirtualSegments(t *testing.T) {
	bus := New()
	bus.Write32(0x00000000, 0x11223344)

	for _, addr := range []uint32{0x00000000, 0x00200000, 0x80000000, 0xA0000000} {
		if got := bus.Read32(addr); got != 0x11223344 {
			t.Fatalf("Read32(0x%08X) = 0x%08X, want 0x11223344", addr, got)
		}
	}
}

func TestScratchpadReadWrite(t *testing.T) {
	bus := New()
	bus.Write16(ScratchpadBase+2, 0xBEEF)

	if got := bus.Read16(ScratchpadBase + 2); got != 0xBEEF {
		t.Fatalf("Read16(scratchpad) = 0x%04X, want 0xBEEF", got)
	}
}

func TestLoadBIOSExposesBootROM(t *testing.T) {
	bios := make([]byte, BIOSSize)
	bios[0] = 0x78
	bios[1] = 0x56
	bios[2] = 0x34
	bios[3] = 0x12

	bus, err := NewWithBIOS(bios)
	if err != nil {
		t.Fatalf("NewWithBIOS returned error: %v", err)
	}

	if got := bus.Read32(BIOSBase); got != 0x12345678 {
		t.Fatalf("Read32(BIOSBase) = 0x%08X, want 0x12345678", got)
	}

	if got := bus.Read32(0xBFC00000); got != 0x12345678 {
		t.Fatalf("Read32(kseg1 BIOS) = 0x%08X, want 0x12345678", got)
	}
}

func TestWriteToBIOSIsIgnored(t *testing.T) {
	bus := New()
	bus.Write32(BIOSBase, 0xDEADBEEF)

	if got := bus.Read32(BIOSBase); got != 0 {
		t.Fatalf("Read32(BIOSBase) = 0x%08X, want 0", got)
	}
}

func TestLoadBIOSRejectsWrongSize(t *testing.T) {
	bus := New()
	if err := bus.LoadBIOS(make([]byte, BIOSSize-1)); err != ErrInvalidBIOSSize {
		t.Fatalf("LoadBIOS returned %v, want %v", err, ErrInvalidBIOSSize)
	}
}

func TestDMAMMIOAccess(t *testing.T) {
	bus := New()

	// DPCR register at 0x1F8010F0 / KSEG1 0xBF8010F0
	defaultDPCR := bus.Read32(0xBF8010F0)
	if defaultDPCR != 0x07654321 {
		t.Fatalf("Read32(0xBF8010F0) = 0x%08X, want 0x07654321", defaultDPCR)
	}

	// Write new DPCR value
	bus.Write32(0xBF8010F0, 0x12345678)
	if got := bus.Read32(0x1F8010F0); got != 0x12345678 {
		t.Fatalf("Read32(0x1F8010F0) = 0x%08X, want 0x12345678", got)
	}
}
