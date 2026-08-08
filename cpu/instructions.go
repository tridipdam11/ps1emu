package cpu

func (c *CPU) branchTarget(op instruction) uint32 {
	return c.CurrentPC + 4 + uint32(op.immSE()<<2)
}

func (c *CPU) jumpTarget(op instruction) uint32 {
	return ((c.CurrentPC + 4) & 0xF0000000) | (op.target() << 2)
}

func (c *CPU) effectiveAddr(op instruction) uint32 {
	return c.Reg(op.rs()) + uint32(op.immSE())
}

func (c *CPU) loadMergeValue(index int) uint32 {
	if c.Load.reg == index {
		return c.Load.value
	}

	return c.Reg(index)
}

func (c *CPU) branchAndLink(op instruction) {
	c.SetReg(31, c.CurrentPC+8)
	c.SetBranchTarget(c.branchTarget(op))
}

func (c *CPU) opSLL(op instruction) {
	if op.raw() == 0 {
		c.R[0] = 0
		return
	}

	c.SetReg(op.rd(), c.Reg(op.rt())<<op.sa())
}

func (c *CPU) opSRL(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rt())>>op.sa())
}

func (c *CPU) opSRA(op instruction) {
	c.SetReg(op.rd(), uint32(int32(c.Reg(op.rt()))>>op.sa()))
}

func (c *CPU) opSLLV(op instruction) {
	shift := c.Reg(op.rs()) & 0x1F
	c.SetReg(op.rd(), c.Reg(op.rt())<<shift)
}

func (c *CPU) opSRLV(op instruction) {
	shift := c.Reg(op.rs()) & 0x1F
	c.SetReg(op.rd(), c.Reg(op.rt())>>shift)
}

func (c *CPU) opSRAV(op instruction) {
	shift := c.Reg(op.rs()) & 0x1F
	c.SetReg(op.rd(), uint32(int32(c.Reg(op.rt()))>>shift))
}

func (c *CPU) opJR(op instruction) {
	c.SetBranchTarget(c.Reg(op.rs()))
}

func (c *CPU) opJALR(op instruction) {
	c.SetReg(op.rd(), c.CurrentPC+8)
	c.SetBranchTarget(c.Reg(op.rs()))
}

func (c *CPU) opSYSCALL(op instruction) { c.exceptionSyscall() }
func (c *CPU) opBREAK(op instruction)   { c.exceptionBreak() }
func (c *CPU) opMFHI(op instruction)    { c.SetReg(op.rd(), c.HI) }
func (c *CPU) opMTHI(op instruction)    { c.HI = c.Reg(op.rs()) }
func (c *CPU) opMFLO(op instruction)    { c.SetReg(op.rd(), c.LO) }
func (c *CPU) opMTLO(op instruction)    { c.LO = c.Reg(op.rs()) }

func (c *CPU) opMULT(op instruction) {
	product := int64(int32(c.Reg(op.rs()))) * int64(int32(c.Reg(op.rt())))
	c.LO = uint32(product)
	c.HI = uint32(uint64(product) >> 32)
}

func (c *CPU) opMULTU(op instruction) {
	product := uint64(c.Reg(op.rs())) * uint64(c.Reg(op.rt()))
	c.LO = uint32(product)
	c.HI = uint32(product >> 32)
}

func (c *CPU) opDIV(op instruction) {
	rs := int32(c.Reg(op.rs()))
	rt := int32(c.Reg(op.rt()))

	switch {
	case rt == 0:
		c.HI = uint32(rs)
		if rs >= 0 {
			c.LO = 0xFFFFFFFF
		} else {
			c.LO = 1
		}
	case rs == -2147483648 && rt == -1:
		c.HI = 0
		c.LO = 0x80000000
	default:
		c.LO = uint32(rs / rt)
		c.HI = uint32(rs % rt)
	}
}

func (c *CPU) opDIVU(op instruction) {
	rs := c.Reg(op.rs())
	rt := c.Reg(op.rt())

	if rt == 0 {
		c.HI = rs
		c.LO = 0xFFFFFFFF
		return
	}

	c.LO = rs / rt
	c.HI = rs % rt
}

func (c *CPU) opADD(op instruction) {
	rs := int32(c.Reg(op.rs()))
	rt := int32(c.Reg(op.rt()))
	sum := rs + rt

	if (rs > 0 && rt > 0 && sum < 0) || (rs < 0 && rt < 0 && sum >= 0) {
		c.exceptionOverflow()
		return
	}

	c.SetReg(op.rd(), uint32(sum))
}

func (c *CPU) opADDU(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rs())+c.Reg(op.rt()))
}

func (c *CPU) opSUB(op instruction) {
	rs := int32(c.Reg(op.rs()))
	rt := int32(c.Reg(op.rt()))
	diff := rs - rt

	if (rs >= 0 && rt < 0 && diff < 0) || (rs < 0 && rt > 0 && diff >= 0) {
		c.exceptionOverflow()
		return
	}

	c.SetReg(op.rd(), uint32(diff))
}

func (c *CPU) opSUBU(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rs())-c.Reg(op.rt()))
}

func (c *CPU) opAND(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rs())&c.Reg(op.rt()))
}

func (c *CPU) opOR(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rs())|c.Reg(op.rt()))
}

func (c *CPU) opXOR(op instruction) {
	c.SetReg(op.rd(), c.Reg(op.rs())^c.Reg(op.rt()))
}

func (c *CPU) opNOR(op instruction) {
	c.SetReg(op.rd(), ^(c.Reg(op.rs()) | c.Reg(op.rt())))
}

func (c *CPU) opSLT(op instruction) {
	if int32(c.Reg(op.rs())) < int32(c.Reg(op.rt())) {
		c.SetReg(op.rd(), 1)
		return
	}

	c.SetReg(op.rd(), 0)
}

func (c *CPU) opSLTU(op instruction) {
	if c.Reg(op.rs()) < c.Reg(op.rt()) {
		c.SetReg(op.rd(), 1)
		return
	}

	c.SetReg(op.rd(), 0)
}

func (c *CPU) opJ(op instruction) {
	c.SetBranchTarget(c.jumpTarget(op))
}

func (c *CPU) opJAL(op instruction) {
	c.SetReg(31, c.CurrentPC+8)
	c.SetBranchTarget(c.jumpTarget(op))
}

func (c *CPU) opBEQ(op instruction) {
	if c.Reg(op.rs()) == c.Reg(op.rt()) {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opBNE(op instruction) {
	if c.Reg(op.rs()) != c.Reg(op.rt()) {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opBLEZ(op instruction) {
	if int32(c.Reg(op.rs())) <= 0 {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opBGTZ(op instruction) {
	if int32(c.Reg(op.rs())) > 0 {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opADDI(op instruction) {
	rs := int32(c.Reg(op.rs()))
	imm := op.immSE()
	sum := rs + imm

	if (rs > 0 && imm > 0 && sum < 0) || (rs < 0 && imm < 0 && sum >= 0) {
		c.exceptionOverflow()
		return
	}

	c.SetReg(op.rt(), uint32(sum))
}

func (c *CPU) opADDIU(op instruction) {
	c.SetReg(op.rt(), c.Reg(op.rs())+uint32(op.immSE()))
}

func (c *CPU) opSLTI(op instruction) {
	if int32(c.Reg(op.rs())) < op.immSE() {
		c.SetReg(op.rt(), 1)
		return
	}

	c.SetReg(op.rt(), 0)
}

func (c *CPU) opSLTIU(op instruction) {
	if c.Reg(op.rs()) < uint32(op.immSE()) {
		c.SetReg(op.rt(), 1)
		return
	}

	c.SetReg(op.rt(), 0)
}

func (c *CPU) opANDI(op instruction) {
	c.SetReg(op.rt(), c.Reg(op.rs())&op.immZE())
}

func (c *CPU) opORI(op instruction) {
	c.SetReg(op.rt(), c.Reg(op.rs())|op.immZE())
}

func (c *CPU) opXORI(op instruction) {
	c.SetReg(op.rt(), c.Reg(op.rs())^op.immZE())
}

func (c *CPU) opLUI(op instruction) {
	c.SetReg(op.rt(), op.immZE()<<16)
}

func (c *CPU) opBLTZ(op instruction) {
	if int32(c.Reg(op.rs())) < 0 {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opBGEZ(op instruction) {
	if int32(c.Reg(op.rs())) >= 0 {
		c.SetBranchTarget(c.branchTarget(op))
	}
}

func (c *CPU) opBLTZAL(op instruction) {
	if int32(c.Reg(op.rs())) < 0 {
		c.branchAndLink(op)
	}
}

func (c *CPU) opBGEZAL(op instruction) {
	if int32(c.Reg(op.rs())) >= 0 {
		c.branchAndLink(op)
	}
}

func (c *CPU) opLB(op instruction) {
	addr := c.effectiveAddr(op)
	value := uint32(int32(int8(c.Read8(addr))))
	c.QueueLoad(op.rt(), value)
}

func (c *CPU) opLH(op instruction) {
	addr := c.effectiveAddr(op)
	if addr&0x1 != 0 {
		c.exceptionAddressLoad(addr)
		return
	}

	value := uint32(int32(int16(c.Read16(addr))))
	c.QueueLoad(op.rt(), value)
}

func (c *CPU) opLWL(op instruction) {
	addr := c.effectiveAddr(op)
	aligned := addr &^ 0x3
	word := c.Read32(aligned)
	value := c.loadMergeValue(op.rt())

	switch addr & 0x3 {
	case 0:
		value = (value & 0x00FFFFFF) | (word << 24)
	case 1:
		value = (value & 0x0000FFFF) | (word << 16)
	case 2:
		value = (value & 0x000000FF) | (word << 8)
	case 3:
		value = word
	}

	c.QueueLoad(op.rt(), value)
}

func (c *CPU) opLW(op instruction) {
	addr := c.effectiveAddr(op)
	if addr&0x3 != 0 {
		c.exceptionAddressLoad(addr)
		return
	}

	c.QueueLoad(op.rt(), c.Read32(addr))
}

func (c *CPU) opLBU(op instruction) {
	addr := c.effectiveAddr(op)
	c.QueueLoad(op.rt(), uint32(c.Read8(addr)))
}

func (c *CPU) opLHU(op instruction) {
	addr := c.effectiveAddr(op)
	if addr&0x1 != 0 {
		c.exceptionAddressLoad(addr)
		return
	}

	c.QueueLoad(op.rt(), uint32(c.Read16(addr)))
}

func (c *CPU) opLWR(op instruction) {
	addr := c.effectiveAddr(op)
	aligned := addr &^ 0x3
	word := c.Read32(aligned)
	value := c.loadMergeValue(op.rt())

	switch addr & 0x3 {
	case 0:
		value = word
	case 1:
		value = (value & 0xFF000000) | (word >> 8)
	case 2:
		value = (value & 0xFFFF0000) | (word >> 16)
	case 3:
		value = (value & 0xFFFFFF00) | (word >> 24)
	}

	c.QueueLoad(op.rt(), value)
}

func (c *CPU) opSB(op instruction) {
	addr := c.effectiveAddr(op)
	c.Write8(addr, uint8(c.Reg(op.rt())))
}

func (c *CPU) opSH(op instruction) {
	addr := c.effectiveAddr(op)
	if addr&0x1 != 0 {
		c.exceptionAddressStore(addr)
		return
	}

	c.Write16(addr, uint16(c.Reg(op.rt())))
}

func (c *CPU) opSWL(op instruction) {
	addr := c.effectiveAddr(op)
	aligned := addr &^ 0x3
	word := c.Read32(aligned)
	value := c.Reg(op.rt())

	switch addr & 0x3 {
	case 0:
		word = (word & 0xFFFFFF00) | (value >> 24)
	case 1:
		word = (word & 0xFFFF0000) | (value >> 16)
	case 2:
		word = (word & 0xFF000000) | (value >> 8)
	case 3:
		word = value
	}

	c.Write32(aligned, word)
}

func (c *CPU) opSW(op instruction) {
	addr := c.effectiveAddr(op)
	if addr&0x3 != 0 {
		c.exceptionAddressStore(addr)
		return
	}

	c.Write32(addr, c.Reg(op.rt()))
}

func (c *CPU) opSWR(op instruction) {
	addr := c.effectiveAddr(op)
	aligned := addr &^ 0x3
	word := c.Read32(aligned)
	value := c.Reg(op.rt())

	switch addr & 0x3 {
	case 0:
		word = value
	case 1:
		word = (word & 0x000000FF) | (value << 8)
	case 2:
		word = (word & 0x0000FFFF) | (value << 16)
	case 3:
		word = (word & 0x00FFFFFF) | (value << 24)
	}

	c.Write32(aligned, word)
}
