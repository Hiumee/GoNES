package internals

type APU struct{}

func (apu *APU) ReadRegister(address uint16) uint8 {
	panic("Not implemented")
}

func (apu *APU) WriteRegister(address uint16, value uint8) {
	panic("Not implemented")
}
