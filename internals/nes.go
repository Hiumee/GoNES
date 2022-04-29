package internals

import (
	"os"
)

type NES struct {
	CPU         *CPU
	APU         *APU
	PPU         *PPU
	Cartridge   *Cartridge
	Bus         *Bus
	Controllers [2]Controller
	RAM         [0x2000]uint8
}

func NewNES() *NES {
	// TODO: init all components
	var nes NES
	var bus *Bus = &Bus{}
	var cpu *CPU = &CPU{}
	var ppu *PPU = &PPU{}
	cpu.PowerUp()
	bus.nes = &nes
	cpu.Bus = bus
	nes.CPU = cpu
	nes.PPU = ppu
	nes.Cartridge = &Cartridge{}

	nes.Initialize()

	return &nes
}

func (nes *NES) LoadFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic("Could not read the input file")
	}

	if data[0] != 'N' || data[1] != 'E' || data[2] != 'S' || data[3] != 0x1A {
		panic("Invalid file format")
	}

	nes.Cartridge.Header.PRG_ROM_size = uint(data[4]) * 16 * 1024
	nes.Cartridge.Header.CHR_ROM_size = uint(data[5]) * 8 * 1024

	// Flags 6
	/*
		Mirroring
		Battery RAM
		Trainer
		Ignore mirroring
		Lower nibble of Mapper #
	*/
	flags6 := data[6]
	nes.Cartridge.Header.Mirroring = flags6&(0x1<<0) == 1
	nes.Cartridge.Header.PersistentRAM = flags6&(0x1<<1) == 1
	nes.Cartridge.Header.Trainer = flags6&(0x1<<2) == 1
	nes.Cartridge.Header.IgnoreMorriring = flags6&(0x1<<3) == 1
	nes.Cartridge.Header.Mapper = uint(flags6&0xF0) >> 4

	// Flags 7
	/*
		VS Unisystem
		PlayChoice - Ignore
		NES2.0 Format - Ignore [2-3]
		Upper nibble of Mapper #
	*/
	flags7 := data[7]
	nes.Cartridge.Header.VSUnisystem = flags7&(0x1<<0) == 1
	nes.Cartridge.Header.Mapper = uint(flags7&0xF0) | nes.Cartridge.Header.Mapper

	nes.Cartridge.Header.PRG_RAM_size = uint(data[8]) * 8 * 1024

	nes.Cartridge.PRG_ROM = make([]byte, nes.Cartridge.Header.PRG_ROM_size)
	nes.Cartridge.CHR_ROM = make([]byte, nes.Cartridge.Header.CHR_ROM_size)

	nes.Cartridge.Loaded = true

	var pointer uint = 16
	// TODO: Trainer?
	memcpy(nes.Cartridge.PRG_ROM, data[pointer:(pointer+nes.Cartridge.Header.PRG_ROM_size)], nes.Cartridge.Header.PRG_ROM_size)
	pointer += nes.Cartridge.Header.PRG_ROM_size
	memcpy(nes.Cartridge.CHR_ROM, data[pointer:(pointer+nes.Cartridge.Header.CHR_ROM_size)], nes.Cartridge.Header.CHR_ROM_size)
	pointer += nes.Cartridge.Header.CHR_ROM_size

	nes.Initialize()
}

func (nes *NES) Initialize() {
	nes.CPU.PowerUp()
	nes.PPU.Initialize()
	//nes.APU.Initialize()
}

func (nes *NES) Step() uint64 {
	// TODO: step all components
	var cycles uint64

	// For each PPU cycle, there are 3 CPU cycles at the same time
	nes.CPU.Cycle()
	nes.CPU.Cycle()
	nes.CPU.Cycle()
	nes.PPU.Cycle()

	return cycles
}
