package cpu

const (
	NumGPR       = 32
	ResetVector  = 0xBFC00000
	BadLoadReg   = -1
	Cop0RegCount = 32
	ExceptionVec = exceptionVector
)

type Bus interface {
	Read8(addr uint32) uint8
	Read16(addr uint32) uint16
	Read32(addr uint32) uint32
	Write8(addr uint32, v uint8)
	Write16(addr uint32, v uint16)
	Write32(addr uint32, v uint32)
}

type pendingLoad struct {
	reg   int
	value uint32
}

type CPU struct {
	R  [NumGPR]uint32
	HI uint32
	LO uint32

	PC     uint32
	NextPC uint32

	CurrentPC uint32

	Cop0 [Cop0RegCount]uint32

	Bus Bus

	Load     pendingLoad
	NextLoad pendingLoad

	InDelaySlot     bool
	NextDelaySlot   bool
	BranchTaken     bool
	BranchTarget    uint32
	LastInstruction uint32
	ExceptionRaised bool

	Cycles uint64
}

func New(bus Bus) *CPU {
	c := &CPU{}
	c.AttachBus(bus)
	c.Reset()
	return c
}

func (c *CPU) Reset() {
	for i := range c.R {
		c.R[i] = 0
	}

	for i := range c.Cop0 {
		c.Cop0[i] = 0
	}

	c.HI = 0
	c.LO = 0

	c.PC = ResetVector
	c.NextPC = ResetVector + 4
	c.CurrentPC = ResetVector

	c.Load = pendingLoad{reg: BadLoadReg}
	c.NextLoad = pendingLoad{reg: BadLoadReg}

	c.InDelaySlot = false
	c.NextDelaySlot = false
	c.BranchTaken = false
	c.BranchTarget = 0
	c.LastInstruction = 0
	c.ExceptionRaised = false
	c.Cycles = 0

	c.R[0] = 0
}

func (c *CPU) AttachBus(bus Bus) {
	c.Bus = bus
}

func (c *CPU) SetPC(addr uint32) {
	c.PC = addr
	c.NextPC = addr + 4
	c.CurrentPC = addr
	c.Load = pendingLoad{reg: BadLoadReg}
	c.NextLoad = pendingLoad{reg: BadLoadReg}
	c.InDelaySlot = false
	c.NextDelaySlot = false
	c.BranchTaken = false
	c.BranchTarget = 0
	c.ExceptionRaised = false
}

func (c *CPU) Reg(index int) uint32 {
	if index <= 0 || index >= NumGPR {
		return 0
	}

	return c.R[index]
}

func (c *CPU) SetReg(index int, value uint32) {
	if index <= 0 || index >= NumGPR {
		return
	}

	c.CancelLoad(index)
	c.R[index] = value
}

func (c *CPU) AdvancePC() {
	c.CurrentPC = c.PC
	c.PC = c.NextPC
	c.NextPC += 4
}

func (c *CPU) SetBranchTarget(addr uint32) {
	c.BranchTaken = true
	c.BranchTarget = addr
	c.NextDelaySlot = true
}

func (c *CPU) ApplyBranch() {
	if c.BranchTaken {
		c.NextPC = c.BranchTarget
	}

	c.InDelaySlot = c.NextDelaySlot
	c.NextDelaySlot = false
	c.BranchTaken = false
	c.BranchTarget = 0
}

func (c *CPU) Fetch32(addr uint32) uint32 {
	if c.Bus == nil {
		return 0
	}

	return c.Bus.Read32(addr)
}

func (c *CPU) Read8(addr uint32) uint8 {
	if c.Bus == nil {
		return 0
	}

	return c.Bus.Read8(addr)
}

func (c *CPU) Read16(addr uint32) uint16 {
	if c.Bus == nil {
		return 0
	}

	return c.Bus.Read16(addr)
}

func (c *CPU) Read32(addr uint32) uint32 {
	if c.Bus == nil {
		return 0
	}

	return c.Bus.Read32(addr)
}

func (c *CPU) Write8(addr uint32, value uint8) {
	if c.Bus == nil {
		return
	}

	c.Bus.Write8(addr, value)
}

func (c *CPU) Write16(addr uint32, value uint16) {
	if c.Bus == nil {
		return
	}

	c.Bus.Write16(addr, value)
}

func (c *CPU) Write32(addr uint32, value uint32) {
	if c.Bus == nil {
		return
	}

	c.Bus.Write32(addr, value)
}

func (c *CPU) SnapshotRegs() [NumGPR]uint32 {
	var out [NumGPR]uint32
	copy(out[:], c.R[:])
	out[0] = 0
	return out
}
