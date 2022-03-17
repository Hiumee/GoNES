package internals

type PPU struct{}

func (ppu *PPU) ReadRegister(address uint16) uint8 {
	panic("Not implemented")
}

func (ppu *PPU) WriteRegister(address uint16, value uint8) {
	panic("Not implemented")
}
