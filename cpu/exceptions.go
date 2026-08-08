package cpu

const (
	cop0BadVAddr = 8

	excAddrLoad    = 4
	excAddrStore   = 5
	excSyscall     = 8
	excBreak       = 9
	excReserved    = 10
	excCopUnusable = 11
	excOverflow    = 12

	causeExcCodeShift = 2
	causeExcCodeMask  = 0x1F << causeExcCodeShift
	causeCopShift     = 28
	causeCopMask      = 0x3 << causeCopShift
	causeBD           = 1 << 31

	srModeMask = 0x3F
	srUserMode = 0x2
	srBEV      = 1 << 22

	exceptionVector     = 0x80000080
	bootExceptionVector = 0xBFC00180
)

const (
	excCodeAdEL = excAddrLoad
	excCodeAdES = excAddrStore
	excCodeSys  = excSyscall
	excCodeBp   = excBreak
	excCodeRI   = excReserved
	excCodeCpU  = excCopUnusable
	excCodeOv   = excOverflow
)

func (c *CPU) enterException(code uint32, coprocessor uint32, badVAddr uint32, setBadVAddr bool) {
	c.ExceptionRaised = true

	if setBadVAddr {
		c.Cop0[cop0BadVAddr] = badVAddr
	}

	cause := c.Cop0[cop0Cause] &^ (causeExcCodeMask | causeCopMask | causeBD)
	cause |= (code << causeExcCodeShift) & causeExcCodeMask
	cause |= (coprocessor << causeCopShift) & causeCopMask

	epc := c.CurrentPC
	if c.InDelaySlot {
		cause |= causeBD
		epc -= instructionSize
	}

	c.Cop0[cop0Cause] = cause
	c.Cop0[cop0EPC] = epc

	sr := c.Cop0[cop0SR]
	c.Cop0[cop0SR] = (sr &^ srModeMask) | ((sr << 2) & srModeMask)

	vector := uint32(exceptionVector)
	if sr&srBEV != 0 {
		vector = uint32(bootExceptionVector)
	}

	c.NextLoad = pendingLoad{reg: BadLoadReg}
	c.BranchTaken = false
	c.BranchTarget = 0
	c.NextDelaySlot = false
	c.InDelaySlot = false
	c.PC = vector
	c.NextPC = vector + uint32(instructionSize)
}

func (c *CPU) exceptionAddressLoad(addr uint32) {
	c.enterException(excAddrLoad, 0, addr, true)
}

func (c *CPU) exceptionAddressStore(addr uint32) {
	c.enterException(excAddrStore, 0, addr, true)
}

func (c *CPU) exceptionSyscall() {
	c.enterException(excSyscall, 0, 0, false)
}

func (c *CPU) exceptionBreak() {
	c.enterException(excBreak, 0, 0, false)
}

func (c *CPU) exceptionReservedInstruction() {
	c.enterException(excReserved, 0, 0, false)
}

func (c *CPU) exceptionCoprocessorUnusable(coprocessor uint32) {
	c.enterException(excCopUnusable, coprocessor, 0, false)
}

func (c *CPU) exceptionOverflow() {
	c.enterException(excOverflow, 0, 0, false)
}

func (c *CPU) inUserMode() bool {
	return c.Cop0[cop0SR]&srUserMode != 0
}
