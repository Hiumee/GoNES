package internals

type Header struct {
	PRG_ROM_size    uint
	CHR_ROM_size    uint
	PRG_RAM_size    uint
	Mirroring       bool
	PersistentRAM   bool // 0x6000 - 0x7FFF
	Trainer         bool // 0x7000 - 0x71FF
	IgnoreMorriring bool
	Mapper          uint
	VSUnisystem     bool
}

type Cartridge struct {
	Loaded  bool
	Header  Header
	PRG_ROM []byte
	CHR_ROM []byte
	RAM     [0x2000]byte
}

func (cartridge *Cartridge) Read(address uint16) uint8 {
	switch {
	case address < 0x2000: // Used for the PPU bus
		return cartridge.CHR_ROM[address]
	case address < 0x8000: // Used for the CPU bus
		return cartridge.RAM[address-0x6000]
	case address >= 0x8000: // Used for the CPU bus
		return cartridge.PRG_ROM[(address-0x8000)%uint16(len(cartridge.PRG_ROM))]
	default:
		return 0
	}
}

func (cartridge *Cartridge) Write(address uint16, value uint8) {
	switch {
	case address < 0x2000: // Used for the PPU bus
		cartridge.CHR_ROM[address] = value
	case address < 0x8000: // Used for the CPU bus
		cartridge.RAM[address-0x6000] = value
	}
}
