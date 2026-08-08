package cpu

import "testing"

type testBus struct {
	mem32 map[uint32]uint32
}

func newTestBus() *testBus {
	return &testBus{
		mem32: make(map[uint32]uint32),
	}
}

func (b *testBus) Read8(addr uint32) uint8 {
	word := b.Read32(addr &^ 0x3)
	shift := (addr & 0x3) * 8
	return uint8(word >> shift)
}

func (b *testBus) Read16(addr uint32) uint16 {
	word := b.Read32(addr &^ 0x3)
	shift := (addr & 0x2) * 8
	return uint16(word >> shift)
}

func (b *testBus) Read32(addr uint32) uint32 {
	return b.mem32[addr]
}

func (b *testBus) Write8(addr uint32, v uint8) {
	base := addr &^ 0x3
	word := b.mem32[base]
	shift := (addr & 0x3) * 8
	mask := uint32(0xFF) << shift
	word = (word &^ mask) | (uint32(v) << shift)
	b.mem32[base] = word
}

func (b *testBus) Write16(addr uint32, v uint16) {
	base := addr &^ 0x3
	word := b.mem32[base]
	shift := (addr & 0x2) * 8
	mask := uint32(0xFFFF) << shift
	word = (word &^ mask) | (uint32(v) << shift)
	b.mem32[base] = word
}

func (b *testBus) Write32(addr uint32, v uint32) {
	b.mem32[addr] = v
}

func assertException(t *testing.T, c *CPU, code uint32, epc uint32, cause uint32) {
	t.Helper()

	if got := c.Cop0[cop0Cause]; got != cause {
		t.Fatalf("Cause = 0x%08X, want 0x%08X", got, cause)
	}

	if got := c.Cop0[cop0EPC]; got != epc {
		t.Fatalf("EPC = 0x%08X, want 0x%08X", got, epc)
	}

	if got := (c.Cop0[cop0Cause] >> 2) & 0x1F; got != code {
		t.Fatalf("ExcCode = %d, want %d", got, code)
	}

	if c.PC != ExceptionVec {
		t.Fatalf("PC = 0x%08X, want 0x%08X", c.PC, uint32(ExceptionVec))
	}

	if c.NextPC != ExceptionVec+4 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, uint32(ExceptionVec+4))
	}
}

func TestNewCPUStartsAtResetVector(t *testing.T) {
	bus := newTestBus()
	c := New(bus)

	if c.PC != ResetVector {
		t.Fatalf("PC = 0x%08X, want 0x%08X", c.PC, ResetVector)
	}

	if c.NextPC != ResetVector+4 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+4)
	}

	if c.Bus != bus {
		t.Fatal("CPU did not keep the attached bus")
	}
}

func TestStepNOPAdvancesCPUState(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00000000)

	c := New(bus)
	result := c.StepResult()

	if result.PC != ResetVector {
		t.Fatalf("result.PC = 0x%08X, want 0x%08X", result.PC, ResetVector)
	}

	if result.Instruction != 0x00000000 {
		t.Fatalf("instruction = 0x%08X, want NOP", result.Instruction)
	}

	if result.Cycles != 1 {
		t.Fatalf("cycles = %d, want 1", result.Cycles)
	}

	if c.PC != ResetVector+4 {
		t.Fatalf("PC = 0x%08X, want 0x%08X", c.PC, ResetVector+4)
	}

	if c.NextPC != ResetVector+8 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+8)
	}

	if c.LastInstruction != 0x00000000 {
		t.Fatalf("LastInstruction = 0x%08X, want NOP", c.LastInstruction)
	}

	if c.Cycles != 1 {
		t.Fatalf("total cycles = %d, want 1", c.Cycles)
	}
}

func TestZeroRegisterStaysZero(t *testing.T) {
	bus := newTestBus()
	c := New(bus)

	c.SetReg(0, 1234)
	if c.Reg(0) != 0 {
		t.Fatalf("R0 changed after SetReg: got %d, want 0", c.Reg(0))
	}

	c.QueueLoad(0, 5678)
	c.CommitLoad()
	if c.Reg(0) != 0 {
		t.Fatalf("R0 changed after queued load: got %d, want 0", c.Reg(0))
	}
}

func TestUnalignedFetchRaisesAddressError(t *testing.T) {
	bus := newTestBus()
	c := New(bus)
	c.SetPC(ResetVector + 2)

	c.StepResult()

	assertException(t, c, excCodeAdEL, ResetVector+2, excCodeAdEL<<2)

	if got := c.Cop0[cop0BadVAddr]; got != ResetVector+2 {
		t.Fatalf("BadVAddr = 0x%08X, want 0x%08X", got, uint32(ResetVector+2))
	}
}

func TestStepORIWritesRegister(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x34021234)

	c := New(bus)
	c.StepResult()

	if got := c.Reg(2); got != 0x1234 {
		t.Fatalf("R2 = 0x%08X, want 0x00001234", got)
	}
}

func TestStepADDUCombinesRegisters(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00221821)

	c := New(bus)
	c.SetReg(1, 5)
	c.SetReg(2, 7)

	c.StepResult()

	if got := c.Reg(3); got != 12 {
		t.Fatalf("R3 = %d, want 12", got)
	}
}

func TestStepLUILoadsUpperImmediate(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x3C041234)

	c := New(bus)
	c.StepResult()

	if got := c.Reg(4); got != 0x12340000 {
		t.Fatalf("R4 = 0x%08X, want 0x12340000", got)
	}
}

func TestStepSLLShiftsRegister(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00011080)

	c := New(bus)
	c.SetReg(1, 3)

	c.StepResult()

	if got := c.Reg(2); got != 12 {
		t.Fatalf("R2 = %d, want 12", got)
	}
}

func TestStepBEQSchedulesBranchAfterDelaySlot(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x10220002)
	bus.Write32(ResetVector+4, 0x34030001)
	bus.Write32(ResetVector+12, 0x34040002)

	c := New(bus)
	c.SetReg(1, 7)
	c.SetReg(2, 7)

	c.StepResult()

	if c.PC != ResetVector+4 {
		t.Fatalf("PC after branch = 0x%08X, want 0x%08X", c.PC, ResetVector+4)
	}

	if c.NextPC != ResetVector+12 {
		t.Fatalf("NextPC after branch = 0x%08X, want 0x%08X", c.NextPC, ResetVector+12)
	}

	if !c.InDelaySlot {
		t.Fatal("expected CPU to enter delay slot after taken branch")
	}

	c.StepResult()

	if got := c.Reg(3); got != 1 {
		t.Fatalf("R3 after delay slot = %d, want 1", got)
	}

	if c.PC != ResetVector+12 {
		t.Fatalf("PC after delay slot = 0x%08X, want 0x%08X", c.PC, ResetVector+12)
	}

	if c.InDelaySlot {
		t.Fatal("expected CPU to leave delay slot after executing branch delay instruction")
	}
}

func TestStepJALStoresReturnAddress(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x0C000004)

	c := New(bus)
	c.StepResult()

	if got := c.Reg(31); got != ResetVector+8 {
		t.Fatalf("RA = 0x%08X, want 0x%08X", got, ResetVector+8)
	}

	if c.NextPC != 0xB0000010 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, uint32(0xB0000010))
	}
}

func TestStepJRUsesRegisterTarget(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00200008)

	c := New(bus)
	c.SetReg(1, ResetVector+0x40)

	c.StepResult()

	if c.NextPC != ResetVector+0x40 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+0x40)
	}
}

func TestStepJALRWritesExplicitDestinationAndSchedulesTarget(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00201809)

	c := New(bus)
	c.SetReg(1, ResetVector+0x40)

	c.StepResult()

	if got := c.Reg(3); got != ResetVector+8 {
		t.Fatalf("R3 = 0x%08X, want 0x%08X", got, ResetVector+8)
	}

	if got := c.Reg(31); got != 0 {
		t.Fatalf("RA = 0x%08X, want 0x00000000", got)
	}

	if c.NextPC != ResetVector+0x40 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+0x40)
	}
}

func TestStepJALRWithZeroDestinationDoesNotWriteRA(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00200009)

	c := New(bus)
	c.SetReg(1, ResetVector+0x40)

	c.StepResult()

	if got := c.Reg(31); got != 0 {
		t.Fatalf("RA = 0x%08X, want 0x00000000", got)
	}
}

func TestStepADDIWritesRegister(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x2022FFFE)

	c := New(bus)
	c.SetReg(1, 5)

	c.StepResult()

	if got := c.Reg(2); got != 3 {
		t.Fatalf("R2 = %d, want 3", got)
	}
}

func TestStepSLTIWritesOneWhenSignedLessThan(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x2822FFFF)

	c := New(bus)
	c.SetReg(1, 0xFFFFFFFE)

	c.StepResult()

	if got := c.Reg(2); got != 1 {
		t.Fatalf("R2 = %d, want 1", got)
	}
}

func TestStepSLTIUUsesUnsignedComparison(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x2C220001)

	c := New(bus)
	c.SetReg(1, 0)

	c.StepResult()

	if got := c.Reg(2); got != 1 {
		t.Fatalf("R2 = %d, want 1", got)
	}
}

func TestStepSLTHandlesSignedEdgeCase(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x0022182A)

	c := New(bus)
	c.SetReg(1, 0x80000000)
	c.SetReg(2, 0x7FFFFFFF)

	c.StepResult()

	if got := c.Reg(3); got != 1 {
		t.Fatalf("R3 = %d, want 1", got)
	}
}

func TestStepSLTUHandlesUnsignedEdgeCase(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x0022182B)

	c := New(bus)
	c.SetReg(1, 0xFFFFFFFF)
	c.SetReg(2, 0x7FFFFFFF)

	c.StepResult()

	if got := c.Reg(3); got != 0 {
		t.Fatalf("R3 = %d, want 0", got)
	}
}

func TestStepSUBSubtractsRegisters(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00221822)

	c := New(bus)
	c.SetReg(1, 9)
	c.SetReg(2, 4)

	c.StepResult()

	if got := c.Reg(3); got != 5 {
		t.Fatalf("R3 = %d, want 5", got)
	}
}

func TestStepADDOverflowRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00221820)

	c := New(bus)
	c.SetReg(1, 0x7FFFFFFF)
	c.SetReg(2, 1)

	c.StepResult()

	assertException(t, c, excCodeOv, ResetVector, excCodeOv<<2)

	if got := c.Reg(3); got != 0 {
		t.Fatalf("R3 = 0x%08X, want 0x00000000", got)
	}
}

func TestStepADDIOverflowRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x20220001)

	c := New(bus)
	c.SetReg(1, 0x7FFFFFFF)

	c.StepResult()

	assertException(t, c, excCodeOv, ResetVector, excCodeOv<<2)

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 = 0x%08X, want 0x00000000", got)
	}
}

func TestStepMULTWritesHILO(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x00220018)

	c := New(bus)
	c.SetReg(1, 0xFFFFFFFE)
	c.SetReg(2, 3)

	c.StepResult()

	if c.LO != 0xFFFFFFFA {
		t.Fatalf("LO = 0x%08X, want 0xFFFFFFFA", c.LO)
	}

	if c.HI != 0xFFFFFFFF {
		t.Fatalf("HI = 0x%08X, want 0xFFFFFFFF", c.HI)
	}
}

func TestStepDIVWritesHILO(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x0022001A)

	c := New(bus)
	c.SetReg(1, 0xFFFFFFF9)
	c.SetReg(2, 2)

	c.StepResult()

	if c.LO != 0xFFFFFFFD {
		t.Fatalf("LO = 0x%08X, want 0xFFFFFFFD", c.LO)
	}

	if c.HI != 0xFFFFFFFF {
		t.Fatalf("HI = 0x%08X, want 0xFFFFFFFF", c.HI)
	}
}

func TestStepBLTZSchedulesBranch(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x04200002)

	c := New(bus)
	c.SetReg(1, 0xFFFFFFFF)

	c.StepResult()

	if c.NextPC != ResetVector+12 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+12)
	}
}

func TestStepBGEZALStoresReturnAddress(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x04310002)

	c := New(bus)
	c.SetReg(1, 0)

	c.StepResult()

	if got := c.Reg(31); got != ResetVector+8 {
		t.Fatalf("RA = 0x%08X, want 0x%08X", got, ResetVector+8)
	}

	if c.NextPC != ResetVector+12 {
		t.Fatalf("NextPC = 0x%08X, want 0x%08X", c.NextPC, ResetVector+12)
	}
}

func TestStepMTC0WritesCop0Register(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x40826000)

	c := New(bus)
	c.SetReg(2, 0x12345678)

	c.StepResult()

	if got := c.Cop0[cop0SR]; got != 0x12345678 {
		t.Fatalf("Cop0[SR] = 0x%08X, want 0x12345678", got)
	}
}

func TestStepMFC0UsesLoadDelay(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x40026000)
	bus.Write32(ResetVector+4, 0x00000000)

	c := New(bus)
	c.Cop0[cop0SR] = 0x89ABCDEC

	c.StepResult()

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 after MFC0 = 0x%08X, want 0x00000000", got)
	}

	c.StepResult()

	if got := c.Reg(2); got != 0x89ABCDEC {
		t.Fatalf("R2 after delay = 0x%08X, want 0x89ABCDEC", got)
	}
}

func TestStepRFERestoresModeBits(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x42000010)

	c := New(bus)
	c.Cop0[cop0SR] = 0x3C

	c.StepResult()

	if got := c.Cop0[cop0SR]; got != 0x3F {
		t.Fatalf("Cop0[SR] = 0x%08X, want 0x0000003F", got)
	}
}

func TestStepCOP0InUserModeRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x40826000)

	c := New(bus)
	c.Cop0[cop0SR] = 0x2
	c.SetReg(2, 0x12345678)

	c.StepResult()

	assertException(t, c, excCodeCpU, ResetVector, excCodeCpU<<2)

	if got := c.Cop0[cop0Cause]; got != excCodeCpU<<2 {
		t.Fatalf("Cause = 0x%08X, want 0x%08X", got, uint32(excCodeCpU<<2))
	}

	if got := c.Cop0[cop0SR]; got != 0x8 {
		t.Fatalf("Cop0[SR] = 0x%08X, want 0x00000008", got)
	}

	if got := c.Cop0[cop0SR]; got == 0x12345678 {
		t.Fatalf("status register was unexpectedly overwritten: 0x%08X", got)
	}
}

func TestStepSYSCALLInDelaySlotSetsBDAndBranchEPC(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x10000001)
	bus.Write32(ResetVector+4, 0x0000000C)

	c := New(bus)

	c.StepResult()
	c.StepResult()

	assertException(t, c, excCodeSys, ResetVector, (1<<31)|(excCodeSys<<2))
}

func TestStepReservedInstructionRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xFC000000)

	c := New(bus)
	c.StepResult()

	assertException(t, c, excCodeRI, ResetVector, excCodeRI<<2)
}

func TestStepLWUsesLoadDelay(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x8C220004)
	bus.Write32(ResetVector+4, 0x00000000)
	bus.Write32(0x1004, 0xDEADBEEF)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 after LW = 0x%08X, want 0x00000000", got)
	}

	c.StepResult()

	if got := c.Reg(2); got != 0xDEADBEEF {
		t.Fatalf("R2 = 0x%08X, want 0xDEADBEEF", got)
	}
}

func TestStepLBSignExtendsByteAfterDelay(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x80220001)
	bus.Write32(ResetVector+4, 0x00000000)
	bus.Write8(0x1001, 0x80)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 after LB = 0x%08X, want 0x00000000", got)
	}

	c.StepResult()

	if got := c.Reg(2); got != 0xFFFFFF80 {
		t.Fatalf("R2 = 0x%08X, want 0xFFFFFF80", got)
	}
}

func TestStepLHUZeroExtendsHalfwordAfterDelay(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x94220002)
	bus.Write32(ResetVector+4, 0x00000000)
	bus.Write16(0x1002, 0xFEDC)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 after LHU = 0x%08X, want 0x00000000", got)
	}

	c.StepResult()

	if got := c.Reg(2); got != 0x0000FEDC {
		t.Fatalf("R2 = 0x%08X, want 0x0000FEDC", got)
	}
}

func TestStepLWLQueuesMergedValue(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x88220001)
	bus.Write32(ResetVector+4, 0x00000000)
	bus.Write32(0x1000, 0x11223344)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xAABBCCDD)

	c.StepResult()
	c.StepResult()

	if got := c.Reg(2); got != 0x3344CCDD {
		t.Fatalf("R2 = 0x%08X, want 0x3344CCDD", got)
	}
}

func TestStepLWRQueuesMergedValue(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x98220002)
	bus.Write32(ResetVector+4, 0x00000000)
	bus.Write32(0x1000, 0x11223344)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xAABBCCDD)

	c.StepResult()
	c.StepResult()

	if got := c.Reg(2); got != 0xAABB1122 {
		t.Fatalf("R2 = 0x%08X, want 0xAABB1122", got)
	}
}

func TestStepLoadDelayCanBeCancelledByRegisterWrite(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x8C220004)
	bus.Write32(ResetVector+4, 0x34020001)
	bus.Write32(0x1004, 0xDEADBEEF)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()
	c.StepResult()

	if got := c.Reg(2); got != 1 {
		t.Fatalf("R2 = 0x%08X, want 0x00000001", got)
	}
}

func TestStepBackToBackLoadsKeepNewestValue(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x8C220000)
	bus.Write32(ResetVector+4, 0x8C220004)
	bus.Write32(ResetVector+8, 0x00000000)
	bus.Write32(0x1000, 0x11111111)
	bus.Write32(0x1004, 0x22222222)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()
	c.StepResult()

	if got := c.Reg(2); got != 0 {
		t.Fatalf("R2 after second load = 0x%08X, want 0x00000000", got)
	}

	c.StepResult()

	if got := c.Reg(2); got != 0x22222222 {
		t.Fatalf("R2 after third step = 0x%08X, want 0x22222222", got)
	}
}

func TestStepSWStoresWord(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xAC220004)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xCAFEBABE)

	c.StepResult()

	if got := bus.Read32(0x1004); got != 0xCAFEBABE {
		t.Fatalf("mem[0x1004] = 0x%08X, want 0xCAFEBABE", got)
	}
}

func TestStepSBStoresLowByteOnly(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xA0220003)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0x123456AB)

	c.StepResult()

	if got := bus.Read8(0x1003); got != 0xAB {
		t.Fatalf("mem[0x1003] = 0x%02X, want 0xAB", got)
	}
}

func TestStepSWLWritesMergedWord(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xA8220001)
	bus.Write32(0x1000, 0x11223344)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xAABBCCDD)

	c.StepResult()

	if got := bus.Read32(0x1000); got != 0x1122AABB {
		t.Fatalf("mem[0x1000] = 0x%08X, want 0x1122AABB", got)
	}
}

func TestStepSWRWritesMergedWord(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xB8220002)
	bus.Write32(0x1000, 0x11223344)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xAABBCCDD)

	c.StepResult()

	if got := bus.Read32(0x1000); got != 0xCCDD3344 {
		t.Fatalf("mem[0x1000] = 0x%08X, want 0xCCDD3344", got)
	}
}

func TestStepLWUnalignedRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0x8C220002)

	c := New(bus)
	c.SetReg(1, 0x1000)

	c.StepResult()

	assertException(t, c, excCodeAdEL, ResetVector, excCodeAdEL<<2)

	if got := c.Cop0[cop0BadVAddr]; got != 0x1002 {
		t.Fatalf("BadVAddr = 0x%08X, want 0x00001002", got)
	}
}

func TestStepSWUnalignedRaisesException(t *testing.T) {
	bus := newTestBus()
	bus.Write32(ResetVector, 0xAC220002)

	c := New(bus)
	c.SetReg(1, 0x1000)
	c.SetReg(2, 0xCAFEBABE)

	c.StepResult()

	assertException(t, c, excCodeAdES, ResetVector, excCodeAdES<<2)

	if got := c.Cop0[cop0BadVAddr]; got != 0x1002 {
		t.Fatalf("BadVAddr = 0x%08X, want 0x00001002", got)
	}

	if got := bus.Read32(0x1000); got != 0 {
		t.Fatalf("mem[0x1000] = 0x%08X, want 0x00000000", got)
	}
}
