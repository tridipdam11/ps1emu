package cpu

const (
	opSPECIAL = 0x00
	opREGIMM  = 0x01
	opJ       = 0x02
	opJAL     = 0x03
	opBEQ     = 0x04
	opBNE     = 0x05
	opBLEZ    = 0x06
	opBGTZ    = 0x07
	opADDI    = 0x08
	opADDIU   = 0x09
	opSLTI    = 0x0A
	opSLTIU   = 0x0B
	opANDI    = 0x0C
	opORI     = 0x0D
	opXORI    = 0x0E
	opLUI     = 0x0F
	opCOP0    = 0x10
	opCOP1    = 0x11
	opCOP2    = 0x12
	opCOP3    = 0x13
	opLB      = 0x20
	opLH      = 0x21
	opLWL     = 0x22
	opLW      = 0x23
	opLBU     = 0x24
	opLHU     = 0x25
	opLWR     = 0x26
	opSB      = 0x28
	opSH      = 0x29
	opSWL     = 0x2A
	opSW      = 0x2B
	opSWR     = 0x2E
)

const (
	regimmBLTZ   = 0x00
	regimmBGEZ   = 0x01
	regimmBLTZAL = 0x10
	regimmBGEZAL = 0x11
)

const (
	functSLL   = 0x00
	functSRL   = 0x02
	functSRA   = 0x03
	functSLLV  = 0x04
	functSRLV  = 0x06
	functSRAV  = 0x07
	functJR    = 0x08
	functJALR  = 0x09
	functSYSC  = 0x0C
	functBREAK = 0x0D
	functMFHI  = 0x10
	functMTHI  = 0x11
	functMFLO  = 0x12
	functMTLO  = 0x13
	functMULT  = 0x18
	functMULTU = 0x19
	functDIV   = 0x1A
	functDIVU  = 0x1B
	functADD   = 0x20
	functADDU  = 0x21
	functSUB   = 0x22
	functSUBU  = 0x23
	functAND   = 0x24
	functOR    = 0x25
	functXOR   = 0x26
	functNOR   = 0x27
	functSLT   = 0x2A
	functSLTU  = 0x2B
)

const (
	copMoveFrom = 0x00
	copMoveTo   = 0x04
	copControl  = 0x10
	copFunctRFE = 0x10
)

type instruction uint32

func (op instruction) raw() uint32       { return uint32(op) }
func (op instruction) opcode() uint32    { return uint32(op>>26) & 0x3F }
func (op instruction) rs() int           { return int((op >> 21) & 0x1F) }
func (op instruction) rt() int           { return int((op >> 16) & 0x1F) }
func (op instruction) rd() int           { return int((op >> 11) & 0x1F) }
func (op instruction) sa() uint32        { return uint32((op >> 6) & 0x1F) }
func (op instruction) funct() uint32     { return uint32(op) & 0x3F }
func (op instruction) imm() uint16       { return uint16(op) }
func (op instruction) immSE() int32      { return int32(int16(op.imm())) }
func (op instruction) immZE() uint32     { return uint32(op.imm()) }
func (op instruction) target() uint32    { return uint32(op) & 0x03FFFFFF }
func (op instruction) regimmCode() int   { return op.rt() }
func (op instruction) copOpcode() uint32 { return uint32(op>>21) & 0x1F }

func (c *CPU) executeInstruction(raw uint32) {
	op := instruction(raw)

	switch op.opcode() {
	case opSPECIAL:
		c.executeSPECIAL(op)
	case opREGIMM:
		c.executeREGIMM(op)
	case opJ:
		c.opJ(op)
	case opJAL:
		c.opJAL(op)
	case opBEQ:
		c.opBEQ(op)
	case opBNE:
		c.opBNE(op)
	case opBLEZ:
		c.opBLEZ(op)
	case opBGTZ:
		c.opBGTZ(op)
	case opADDI:
		c.opADDI(op)
	case opADDIU:
		c.opADDIU(op)
	case opSLTI:
		c.opSLTI(op)
	case opSLTIU:
		c.opSLTIU(op)
	case opANDI:
		c.opANDI(op)
	case opORI:
		c.opORI(op)
	case opXORI:
		c.opXORI(op)
	case opLUI:
		c.opLUI(op)
	case opCOP0:
		c.executeCOP0(op)
	case opCOP1:
		c.executeCOP1(op)
	case opCOP2:
		c.executeCOP2(op)
	case opCOP3:
		c.executeCOP3(op)
	case opLB:
		c.opLB(op)
	case opLH:
		c.opLH(op)
	case opLWL:
		c.opLWL(op)
	case opLW:
		c.opLW(op)
	case opLBU:
		c.opLBU(op)
	case opLHU:
		c.opLHU(op)
	case opLWR:
		c.opLWR(op)
	case opSB:
		c.opSB(op)
	case opSH:
		c.opSH(op)
	case opSWL:
		c.opSWL(op)
	case opSW:
		c.opSW(op)
	case opSWR:
		c.opSWR(op)
	default:
		c.exceptionReservedInstruction()
	}
}

func (c *CPU) executeSPECIAL(op instruction) {
	switch op.funct() {
	case functSLL:
		c.opSLL(op)
	case functSRL:
		c.opSRL(op)
	case functSRA:
		c.opSRA(op)
	case functSLLV:
		c.opSLLV(op)
	case functSRLV:
		c.opSRLV(op)
	case functSRAV:
		c.opSRAV(op)
	case functJR:
		c.opJR(op)
	case functJALR:
		c.opJALR(op)
	case functSYSC:
		c.opSYSCALL(op)
	case functBREAK:
		c.opBREAK(op)
	case functMFHI:
		c.opMFHI(op)
	case functMTHI:
		c.opMTHI(op)
	case functMFLO:
		c.opMFLO(op)
	case functMTLO:
		c.opMTLO(op)
	case functMULT:
		c.opMULT(op)
	case functMULTU:
		c.opMULTU(op)
	case functDIV:
		c.opDIV(op)
	case functDIVU:
		c.opDIVU(op)
	case functADD:
		c.opADD(op)
	case functADDU:
		c.opADDU(op)
	case functSUB:
		c.opSUB(op)
	case functSUBU:
		c.opSUBU(op)
	case functAND:
		c.opAND(op)
	case functOR:
		c.opOR(op)
	case functXOR:
		c.opXOR(op)
	case functNOR:
		c.opNOR(op)
	case functSLT:
		c.opSLT(op)
	case functSLTU:
		c.opSLTU(op)
	default:
		c.exceptionReservedInstruction()
	}
}

func (c *CPU) executeREGIMM(op instruction) {
	switch op.regimmCode() {
	case regimmBLTZ:
		c.opBLTZ(op)
	case regimmBGEZ:
		c.opBGEZ(op)
	case regimmBLTZAL:
		c.opBLTZAL(op)
	case regimmBGEZAL:
		c.opBGEZAL(op)
	default:
		c.exceptionReservedInstruction()
	}
}

func (c *CPU) executeCOP0(op instruction) {
	if c.inUserMode() {
		c.exceptionCoprocessorUnusable(0)
		return
	}

	switch op.copOpcode() {
	case copMoveFrom:
		c.opMFC0(op)
	case copMoveTo:
		c.opMTC0(op)
	case copControl:
		if op.funct() == copFunctRFE {
			c.opRFE(op)
			return
		}

		c.exceptionReservedInstruction()
	default:
		c.exceptionReservedInstruction()
	}
}

func (c *CPU) executeCOP1(op instruction) {
	c.exceptionCoprocessorUnusable(1)
}

func (c *CPU) executeCOP2(op instruction) {
	c.exceptionCoprocessorUnusable(2)
}

func (c *CPU) executeCOP3(op instruction) {
	c.exceptionCoprocessorUnusable(3)
}
