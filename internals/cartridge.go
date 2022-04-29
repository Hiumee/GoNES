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
}

func (cartridge *Cartridge) Read(address uint16) uint8 {
	panic("Not implemented")
}

func (cartridge *Cartridge) Write(address uint16, value uint8) {
	panic("Not implemented")
}
