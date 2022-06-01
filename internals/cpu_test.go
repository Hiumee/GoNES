package internals

import (
	"os"
	"testing"
)

type BusMock struct {
	RAM [0x16000]uint8
}

func (memoryMock *BusMock) Read(address uint16) uint8 {
	return memoryMock.RAM[address]
}

func (memoryMock *BusMock) ReadAddress(address uint16) uint16 {
	var low uint16 = uint16(memoryMock.Read(address))
	var high uint16 = uint16(memoryMock.Read(address + 1))
	return low | high<<8
}

func (memoryMock *BusMock) Write(address uint16, value uint8) {
	memoryMock.RAM[address] = value
}

func (memoryMock *BusMock) WriteAddress(address uint16, value uint16) {
	var low uint8 = uint8(value & 0xFF)
	var high uint8 = uint8(value >> 8)
	memoryMock.Write(address, low)
	memoryMock.Write(address+1, high)
}

func (memoryMock *BusMock) ReadAddressBug(address uint16) uint16 {
	var low uint16 = uint16(memoryMock.Read(address))
	var high uint16 = uint16(memoryMock.Read((address+1)&0xFF + address&0xFF00))
	return low | high<<8
}

// Used nestest https://github.com/christopherpow/nes-test-roms/blob/master/other/nestest.txt
func TestCPUInstructions(t *testing.T) {
	var data []uint8
	data, err := os.ReadFile("tests/nestest.bin")
	if err != nil {
		t.Error("Cannot open test file")
	}

	memory := &BusMock{}
	copy(memory.RAM[0x8000:0xC000], data[:0x4000])
	copy(memory.RAM[0xC000:0x10000], data[:0x4000])

	var cpu *CPU = &CPU{}
	cpu.Bus = memory
	cpu.PowerUp()
	cpu.PC = 0xC000

	for cpu.CycleCount < 14940 {
		cpu.Cycle()
	}

	errorCode1 := cpu.Bus.Read(0x2)
	errorCode2 := cpu.Bus.Read(0x3)

	if cpu.PC != 0xC6C4 || cpu.A != 0x55 || cpu.Y != 0x53 || cpu.GetFlags() != 0x24 || cpu.SP != 0xF9 || cpu.CycleCount != 14940 || errorCode1 != 0 || errorCode2 != 0 {
		t.Error("Failed CPU instructions test. Fail codes: ", errorCode1, errorCode2)
	}
}

func TestNonMockCPUInstructions(t *testing.T) {
	nes := NewNES()
	nes.LoadFile("tests/nestest.nes")

	nes.CPU.PC = 0xC000

	for nes.CPU.CycleCount < 14940 {
		nes.CPU.Cycle()
	}

	errorCode1 := nes.Bus.Read(0x2)
	errorCode2 := nes.Bus.Read(0x3)

	if errorCode1 != 0 || errorCode2 != 0 {
		t.Error("Failed CPU instructions test. Fail codes: ", errorCode1, errorCode2)
	}
}
