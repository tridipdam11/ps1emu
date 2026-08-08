package dma

// ChannelPort identifies one of the 7 PlayStation DMA channels.
type ChannelPort int

const (
	PortMDECin  ChannelPort = 0 // Macroblock Decoder Input
	PortMDECout ChannelPort = 1 // Macroblock Decoder Output
	PortGPU     ChannelPort = 2 // Graphics Processing Unit
	PortCDROM   ChannelPort = 3 // CD-ROM Drive Interface
	PortSPU     ChannelPort = 4 // Sound Processing Unit
	PortPIO     ChannelPort = 5 // Extension/PIO Port
	PortOTC     ChannelPort = 6 // Ordering Table Clear
)

// SyncMode indicates how the transfer is synchronized/triggered.
type SyncMode uint8

const (
	SyncImmediate  SyncMode = 0 // Direct/Immediate transfer (starts upon CHCR trigger bit)
	SyncRequest    SyncMode = 1 // Sync to Peripheral Request (block-based)
	SyncLinkedList SyncMode = 2 // Linked List transfer (primarily used for GPU)
)

// Direction indicates the transfer direction relative to RAM.
type Direction uint8

const (
	ToRAM   Direction = 0 // Device -> Main RAM
	FromRAM Direction = 1 // Main RAM -> Device
)

// Channel represents the state and registers of an individual DMA channel.
type Channel struct {
	Port ChannelPort

	MADR uint32 // Memory Address Register (24-bit RAM pointer)
	BCR  uint32 // Block Control Register (Word Count / Block Count)
	CHCR uint32 // Control Register
}

// NewChannel creates a initialized DMA channel.
func NewChannel(port ChannelPort) Channel {
	return Channel{
		Port: port,
	}
}

// Direction returns the data transfer direction (ToRAM or FromRAM).
func (c *Channel) Direction() Direction {
	if (c.CHCR & 1) != 0 {
		return FromRAM
	}
	return ToRAM
}

// Step returns the address step increment in bytes (+4 or -4).
func (c *Channel) Step() int32 {
	if ((c.CHCR >> 1) & 1) != 0 {
		return -4
	}
	return 4
}

// SyncMode returns the channel's synchronization mode (0, 1, or 2).
func (c *Channel) SyncMode() SyncMode {
	return SyncMode((c.CHCR >> 9) & 3)
}

// IsActive returns whether a DMA transfer is currently triggered/active on this channel.
func (c *Channel) IsActive() bool {
	sync := c.SyncMode()
	enable := (c.CHCR & (1 << 24)) != 0
	trigger := (c.CHCR & (1 << 28)) != 0

	if sync == SyncImmediate {
		return enable && trigger
	}
	return enable
}

// Finish deactivates the active DMA transfer flags in CHCR.
func (c *Channel) Finish() {
	c.CHCR &= ^(uint32(1) << 24)
	c.CHCR &= ^(uint32(1) << 28)
}

// ReadReg returns the value of a channel register by offset (0x0: MADR, 0x4: BCR, 0x8: CHCR).
func (c *Channel) ReadReg(offset uint32) uint32 {
	switch offset {
	case 0x0:
		return c.MADR
	case 0x4:
		return c.BCR
	case 0x8:
		return c.CHCR
	default:
		return 0
	}
}

// WriteReg sets the value of a channel register by offset.
func (c *Channel) WriteReg(offset uint32, val uint32) {
	switch offset {
	case 0x0:
		c.MADR = val & 0x00FFFFFF
	case 0x4:
		c.BCR = val
	case 0x8:
		c.CHCR = val
	}
}
