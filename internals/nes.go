package internals

import (
	"fmt"
	"os"
)

type NES struct {
	CPU         *CPU
	APU         *APU
	PPU         *PPU
	Cartridge   *Cartridge
	Controllers [2]Controller
	RAM         [0x2000]uint8
}

func NewNES() *NES {
	// TODO: init all components
	var nes NES
	var memory *Memory = &Memory{}
	var cpu *CPU = &CPU{}
	cpu.PowerUp()
	memory.nes = &nes
	cpu.Memory = memory
	nes.CPU = cpu

	return &nes
}

func (nes *NES) LoadFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic("Could not read the input file")
	}
	fmt.Println(data)
}

func (nes *NES) Step() uint64 {
	// TODO: step all components
	var cycles uint64
	cycles, _ = nes.CPU.Step()

	return cycles
}
