package cpu

func (c *CPU) QueueLoad(index int, value uint32) {
	if index <= 0 || index >= NumGPR {
		c.NextLoad = pendingLoad{reg: BadLoadReg}
		return
	}

	c.CancelLoad(index)
	c.NextLoad = pendingLoad{reg: index, value: value}
}

func (c *CPU) CommitLoad() {
	if c.Load.reg > 0 && c.Load.reg < NumGPR {
		c.R[c.Load.reg] = c.Load.value
	}

	c.Load = c.NextLoad
	c.NextLoad = pendingLoad{reg: BadLoadReg}
	c.R[0] = 0
}

func (c *CPU) CancelLoad(index int) {
	if c.Load.reg == index {
		c.Load = pendingLoad{reg: BadLoadReg}
	}
}
