package internals

import (
	"fmt"
	"os"
	"testing"
)

type MemoryMock struct {
	RAM [0x16000]uint8
}

func (memoryMock *MemoryMock) Read(address uint16) uint8 {
	return memoryMock.RAM[address]
}

func (memoryMock *MemoryMock) ReadAddress(address uint16) uint16 {
	var low uint16 = uint16(memoryMock.Read(address))
	var high uint16 = uint16(memoryMock.Read(address + 1))
	return low | high<<8
}

func (memoryMock *MemoryMock) Write(address uint16, value uint8) {
	memoryMock.RAM[address] = value
}

func (memoryMock *MemoryMock) WriteAddress(address uint16, value uint16) {
	var low uint8 = uint8(value & 0xFF)
	var high uint8 = uint8(value >> 8)
	memoryMock.Write(address, low)
	memoryMock.Write(address+1, high)
}

func TestCPUInstructions(t *testing.T) {
	var data []uint8
	data, err := os.ReadFile("tests/nestest.bin")
	if err != nil {
		t.Error("Cannot open test file")
	}

	memory := &MemoryMock{}
	copy(memory.RAM[0x8000:0xC000], data[:0x4000])
	copy(memory.RAM[0xC000:0x10000], data[:0x4000])

	var cpu *CPU = &CPU{}
	cpu.PowerUp()
	cpu.Memory = memory
	cpu.PC = 0xC000

	defer func() {
		err := recover()
		if err != nil {
			fmt.Printf("Panic: %v\n", err)
		}
		errorCode := memory.RAM[0x2]
		errorCode2 := memory.RAM[0x3]
		if errorCode != 0x0 {
			t.Error("Error code 0x2:", errorCode)
		}
		if errorCode != 0x0 {
			t.Error("Error code 0x3:", errorCode2)
		}
	}()

	for cpu.PC != 0x2000 {
		cpu.Step()
	}
}
