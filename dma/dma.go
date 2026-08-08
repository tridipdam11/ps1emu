package dma

const (
	DMABaseAddress = 0x1F801080
	DMABoundEnd    = 0x1F8010F8
)

// MemoryBus provides RAM access for DMA transfers.
type MemoryBus interface {
	Read32(addr uint32) uint32
	Write32(addr uint32, val uint32)
}

// DMA represents the PlayStation 1 7-channel DMA Controller.
type DMA struct {
	Channels [7]Channel
	DPCR     uint32 // DMA Control Register (0x1F8010F0)
	DICR     uint32 // DMA Interrupt Register (0x1F8010F4)
}

// New creates a new DMA Controller with default reset register states.
func New() *DMA {
	dma := &DMA{
		DPCR: 0x07654321, // Default reset state
		DICR: 0,
	}

	for i := 0; i < 7; i++ {
		dma.Channels[i] = NewChannel(ChannelPort(i))
	}

	return dma
}

// Read32 handles MMIO reads for DMA register offsets relative to 0x1F801080.
func (d *DMA) Read32(offset uint32) uint32 {
	if offset < 0x70 {
		channelIdx := offset / 0x10
		regOffset := offset % 0x10
		if channelIdx < 7 {
			return d.Channels[channelIdx].ReadReg(regOffset)
		}
	}

	switch offset {
	case 0x70: // DPCR
		return d.DPCR
	case 0x74: // DICR
		return d.DICR
	default:
		return 0
	}
}

// Write32 handles MMIO writes for DMA register offsets relative to 0x1F801080.
func (d *DMA) Write32(offset uint32, val uint32, bus MemoryBus) {
	if offset < 0x70 {
		channelIdx := offset / 0x10
		regOffset := offset % 0x10
		if channelIdx < 7 {
			d.Channels[channelIdx].WriteReg(regOffset, val)

			// Trigger transfer if channel was activated
			if regOffset == 0x8 && d.Channels[channelIdx].IsActive() {
				d.ExecuteTransfer(ChannelPort(channelIdx), bus)
			}
		}
		return
	}

	switch offset {
	case 0x70: // DPCR
		d.DPCR = val
	case 0x74: // DICR
		// Bits 24..30 are write-1-to-clear flags
		ackFlags := (val >> 24) & 0x7F
		currentFlags := (d.DICR >> 24) & 0x7F
		newFlags := currentFlags &^ ackFlags

		// Keep low bits, update flags
		d.DICR = (val & 0x00FFFFFF) | (newFlags << 24)
		d.updateInterruptMasterFlag()
	}
}

// ExecuteTransfer handles DMA transfers for a specific channel.
func (d *DMA) ExecuteTransfer(port ChannelPort, bus MemoryBus) {
	ch := &d.Channels[port]

	switch ch.SyncMode() {
	case SyncImmediate:
		if port == PortOTC {
			d.transferOTC(ch, bus)
		} else {
			d.transferBlock(ch, bus)
		}
	case SyncRequest:
		d.transferBlock(ch, bus)
	case SyncLinkedList:
		d.transferLinkedList(ch, bus)
	}

	ch.Finish()
	d.raiseInterrupt(port)
}

// transferOTC performs Ordering Table Clear (Channel 6).
func (d *DMA) transferOTC(ch *Channel, bus MemoryBus) {
	words := ch.BCR & 0xFFFF
	if words == 0 {
		words = 0x10000
	}

	addr := ch.MADR & 0x00FFFFFF
	step := uint32(ch.Step()) // usually -4 (0xFFFFFFFC)

	for i := uint32(0); i < words; i++ {
		var val uint32
		if i == words-1 {
			val = 0x00FFFFFF // End of table marker
		} else {
			val = (addr + step) & 0x00FFFFFF
		}

		bus.Write32(addr, val)
		addr = (addr + step) & 0x00FFFFFF
	}

	ch.MADR = addr
}

// transferBlock handles standard direct DMA block transfer.
func (d *DMA) transferBlock(ch *Channel, bus MemoryBus) {
	wordCount := ch.BCR & 0xFFFF
	if wordCount == 0 {
		wordCount = 0x10000
	}

	blockCount := (ch.BCR >> 16) & 0xFFFF
	if blockCount == 0 {
		blockCount = 1
	}

	totalWords := wordCount * blockCount
	addr := ch.MADR & 0x00FFFFFF
	step := uint32(ch.Step())

	for i := uint32(0); i < totalWords; i++ {
		// Place-holder transfer loop; specific hardware read/write hooks will connect here
		addr = (addr + step) & 0x00FFFFFF
	}

	ch.MADR = addr
}

// transferLinkedList processes GPU Linked List DMA (Channel 2).
func (d *DMA) transferLinkedList(ch *Channel, bus MemoryBus) {
	addr := ch.MADR & 0x00FFFFFF

	for {
		header := bus.Read32(addr)
		packetWords := (header >> 24) & 0xFF
		nextAddr := header & 0x00FFFFFF

		// Process packet data words following header node
		for i := uint32(0); i < packetWords; i++ {
			addr = (addr + 4) & 0x00FFFFFF
			// bus.Read32(addr) sent to GPU command processor
		}

		if (header & 0x00FFFFFF) == 0x00FFFFFF || nextAddr == addr {
			break
		}

		addr = nextAddr
	}

	ch.MADR = addr
}

func (d *DMA) raiseInterrupt(port ChannelPort) {
	// Enable flag bit for channel
	mask := uint32(1) << (16 + uint32(port))
	if (d.DICR & mask) != 0 {
		// Set interrupt flag bit
		d.DICR |= (uint32(1) << (24 + uint32(port)))
	}
	d.updateInterruptMasterFlag()
}

func (d *DMA) updateInterruptMasterFlag() {
	force := (d.DICR & (1 << 15)) != 0
	masterEnable := (d.DICR & (1 << 23)) != 0
	enableMask := (d.DICR >> 16) & 0x7F
	flagMask := (d.DICR >> 24) & 0x7F

	irq := force || (masterEnable && (enableMask&flagMask) != 0)
	if irq {
		d.DICR |= (uint32(1) << 31)
	} else {
		d.DICR &= ^(uint32(1) << 31)
	}
}
