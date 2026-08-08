package cpu

const instructionSize = 4

type StepResult struct {
	PC          uint32
	Instruction uint32
	Cycles      uint32
}

func (c *CPU) Step() uint32 {
	result := c.StepResult()
	return result.Cycles
}

func (c *CPU) StepResult() StepResult {
	pc := c.PC
	c.ExceptionRaised = false

	if pc&0x3 != 0 {
		c.CurrentPC = pc
		c.LastInstruction = 0
		c.exceptionAddressLoad(pc)
		return c.finishStep(pc, 0)
	}

	instruction := c.Fetch32(pc)

	c.CurrentPC = pc
	c.LastInstruction = instruction

	c.AdvancePC()
	c.executeInstruction(instruction)
	return c.finishStep(pc, instruction)
}

func (c *CPU) finishStep(pc uint32, instruction uint32) StepResult {
	c.CommitLoad()

	if c.ExceptionRaised {
		c.ExceptionRaised = false
	} else {
		c.ApplyBranch()
	}

	cycles := uint32(1)
	c.Cycles += uint64(cycles)

	return StepResult{
		PC:          pc,
		Instruction: instruction,
		Cycles:      cycles,
	}
}

func (c *CPU) unimplemented(op uint32) {
	c.LastInstruction = op
	c.exceptionReservedInstruction()
}

func (c *CPU) SetNextPC(addr uint32) {
	c.NextPC = addr
}

func (c *CPU) SkipInstruction() {
	c.NextPC += instructionSize
}
