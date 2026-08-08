package memory

import (
	"encoding/binary"
	"errors"

	"ps1emu/dma"
)

const (
	MainRAMSize    = 2 * 1024 * 1024
	MainRAMMirror  = 8 * 1024 * 1024
	ScratchpadBase = 0x1F800000
	ScratchpadSize = 1024
	BIOSBase       = 0x1FC00000
	BIOSSize       = 512 * 1024
)

var ErrInvalidBIOSSize = errors.New("bios must be exactly 512 KiB")

type Bus struct {
	ram        []byte
	scratchpad []byte
	bios       []byte
	DMA        *dma.DMA
}

func New() *Bus {
	return &Bus{
		ram:        make([]byte, MainRAMSize),
		scratchpad: make([]byte, ScratchpadSize),
		bios:       make([]byte, BIOSSize),
		DMA:        dma.New(),
	}
}

func NewWithBIOS(data []byte) (*Bus, error) {
	b := New()
	if err := b.LoadBIOS(data); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bus) LoadBIOS(data []byte) error {
	if len(data) != BIOSSize {
		return ErrInvalidBIOSSize
	}

	copy(b.bios, data)
	return nil
}

func (b *Bus) RAM() []byte {
	return b.ram
}

func (b *Bus) BIOS() []byte {
	return b.bios
}

func (b *Bus) Read8(addr uint32) uint8 {
	space, offset, ok := b.decode(addr)
	if !ok {
		return 0
	}

	return space[offset]
}

func (b *Bus) Read16(addr uint32) uint16 {
	lo := uint16(b.Read8(addr))
	hi := uint16(b.Read8(addr + 1))
	return lo | (hi << 8)
}

func (b *Bus) Read32(addr uint32) uint32 {
	phys := translateAddress(addr)
	if phys >= dma.DMABaseAddress && phys < dma.DMABoundEnd {
		return b.DMA.Read32(phys - dma.DMABaseAddress)
	}

	var buf [4]byte
	buf[0] = b.Read8(addr)
	buf[1] = b.Read8(addr + 1)
	buf[2] = b.Read8(addr + 2)
	buf[3] = b.Read8(addr + 3)
	return binary.LittleEndian.Uint32(buf[:])
}

func (b *Bus) Write8(addr uint32, value uint8) {
	space, offset, ok := b.decodeWritable(addr)
	if !ok {
		return
	}

	space[offset] = value
}

func (b *Bus) Write16(addr uint32, value uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	b.Write8(addr, buf[0])
	b.Write8(addr+1, buf[1])
}

func (b *Bus) Write32(addr uint32, value uint32) {
	phys := translateAddress(addr)
	if phys >= dma.DMABaseAddress && phys < dma.DMABoundEnd {
		b.DMA.Write32(phys-dma.DMABaseAddress, value, b)
		return
	}

	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	b.Write8(addr, buf[0])
	b.Write8(addr+1, buf[1])
	b.Write8(addr+2, buf[2])
	b.Write8(addr+3, buf[3])
}

func (b *Bus) decode(addr uint32) ([]byte, uint32, bool) {
	phys := translateAddress(addr)

	switch {
	case phys < MainRAMMirror:
		return b.ram, phys % MainRAMSize, true
	case phys >= ScratchpadBase && phys < ScratchpadBase+ScratchpadSize:
		return b.scratchpad, phys - ScratchpadBase, true
	case phys >= BIOSBase && phys < BIOSBase+BIOSSize:
		return b.bios, phys - BIOSBase, true
	default:
		return nil, 0, false
	}
}

func (b *Bus) decodeWritable(addr uint32) ([]byte, uint32, bool) {
	phys := translateAddress(addr)

	switch {
	case phys < MainRAMMirror:
		return b.ram, phys % MainRAMSize, true
	case phys >= ScratchpadBase && phys < ScratchpadBase+ScratchpadSize:
		return b.scratchpad, phys - ScratchpadBase, true
	default:
		return nil, 0, false
	}
}

func translateAddress(addr uint32) uint32 {
	switch addr >> 29 {
	case 4:
		return addr & 0x7FFFFFFF
	case 5:
		return addr & 0x1FFFFFFF
	default:
		return addr
	}
}
