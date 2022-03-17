package internals

type IMemory interface {
	Read(addrress uint16) uint8
	ReadAddress(address uint16) uint16
	Write(address uint16, value uint8)
	WriteAddress(address uint16, value uint16)
}

type Memory struct {
	nes *NES
}

// https://wiki.nesdev.org/w/index.php?title=CPU_memory_map
func (memory *Memory) Read(address uint16) uint8 {
	switch {
	case address < 0x2000:
		return memory.nes.RAM[address%0x0800]
	case address < 0x4000:
		return memory.nes.PPU.ReadRegister(0x4000 + address%0x8)
	case address < 0x4014: // Shouldn't read this
		return memory.nes.APU.ReadRegister(address)
	case address == 0x4014:
		return memory.nes.PPU.ReadRegister(address)
	case address == 0x4015:
		return memory.nes.APU.ReadRegister(address)
	case address == 0x4016:
		return memory.nes.Controllers[0].ReadState()
	case address == 0x4017:
		return memory.nes.Controllers[1].ReadState()
	case address < 0x4020:
		return 0
	default:
		return memory.nes.Cartridge.Read(address)
	}
}

func (memory *Memory) ReadAddress(address uint16) uint16 {
	var low uint16 = uint16(memory.Read(address))
	var high uint16 = uint16(memory.Read(address + 1))
	return low | high<<8
}

func (memory *Memory) Write(address uint16, value uint8) {
	switch {
	case address < 0x2000:
		memory.nes.RAM[address%0x0800] = value
	case address < 0x4000:
		memory.nes.PPU.WriteRegister(0x4000+address%0x08, value)
	case address < 0x4014:
		memory.nes.APU.WriteRegister(address, value)
	case address == 0x4014:
		memory.nes.PPU.WriteRegister(address, value)
	case address == 0x4015:
		memory.nes.APU.WriteRegister(address, value)
	case address == 0x4016:
		memory.nes.Controllers[0].WriteState(value)
	case address == 0x4017:
		memory.nes.APU.WriteRegister(address, value)
	case address < 0x4020:
		//0
	default:
		memory.nes.Cartridge.Write(address, value)
	}
}

func (memory *Memory) WriteAddress(address uint16, value uint16) {
	var low uint8 = uint8(value & 0xFF)
	var high uint8 = uint8(value >> 8)
	memory.Write(address, low)
	memory.Write(address+1, high)
}
