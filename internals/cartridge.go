package internals

type Cartridge struct {
}

func (cartridge *Cartridge) Read(address uint16) uint8 {
	panic("Not implemented")
}

func (cartridge *Cartridge) Write(address uint16, value uint8) {
	panic("Not implemented")
}
