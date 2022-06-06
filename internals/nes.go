package internals

import (
	"io/ioutil"
	"strconv"
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
	var nes NES
	var bus *Bus = &Bus{}
	var cpu *CPU = &CPU{}
	var ppu *PPU = &PPU{}
	nes.Bus = bus
	bus.nes = &nes
	cpu.Bus = bus
	ppu.Bus = bus
	nes.CPU = cpu
	nes.PPU = ppu
	nes.Cartridge = &Cartridge{}

	return &nes
}

func (nes *NES) LoadFile(filename string) {
	data, err := ioutil.ReadFile(filename)
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

	if nes.Cartridge.Header.Mapper != 0 {
		panic("Unsupported mapper: " + strconv.Itoa(int(nes.Cartridge.Header.Mapper)))
	}

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
	var cycles uint64

	// For each CPU cycle, there are 3 PPU cycles at the same time
	nes.CPU.Cycle()
	nes.PPU.Cycle()
	nes.PPU.Cycle()
	nes.PPU.Cycle()

	return cycles
}
