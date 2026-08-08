package cpu

const (
	cop0SR    = 12
	cop0Cause = 13
	cop0EPC   = 14
)

func (c *CPU) opMFC0(op instruction) {
	if c.inUserMode() {
		c.exceptionCoprocessorUnusable(0)
		return
	}

	c.QueueLoad(op.rt(), c.Cop0[op.rd()])
}

func (c *CPU) opMTC0(op instruction) {
	if c.inUserMode() {
		c.exceptionCoprocessorUnusable(0)
		return
	}

	c.Cop0[op.rd()] = c.Reg(op.rt())
}

func (c *CPU) opRFE(op instruction) {
	if c.inUserMode() {
		c.exceptionCoprocessorUnusable(0)
		return
	}

	sr := c.Cop0[cop0SR]
	c.Cop0[cop0SR] = (sr &^ 0xF) | ((sr >> 2) & 0xF)
}
