package internals

import (
	"fmt"
)

var COLOR_PALETTE []uint8 = []uint8{84, 84, 84, 0, 30, 116, 8, 16, 144, 48, 0, 136, 68, 0, 100, 92, 0, 48, 84, 4, 0, 60, 24, 0, 32, 42, 0, 8, 58, 0, 0, 64, 0, 0, 60, 0, 0, 50, 60, 0, 0, 0, 0, 0, 0, 0, 0, 0, 152, 150, 152, 8, 76, 196, 48, 50, 236, 92, 30, 228, 136, 20, 176, 160, 20, 100, 152, 34, 32, 120, 60, 0, 84, 90, 0, 40, 114, 0, 8, 124, 0, 0, 118, 40, 0, 102, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 236, 238, 236, 76, 154, 236, 120, 124, 236, 176, 98, 236, 228, 84, 236, 236, 88, 180, 236, 106, 100, 212, 136, 32, 160, 170, 0, 116, 196, 0, 76, 208, 32, 56, 204, 108, 56, 180, 204, 60, 60, 60, 0, 0, 0, 0, 0, 0, 236, 238, 236, 168, 204, 236, 188, 188, 236, 212, 178, 236, 236, 174, 236, 236, 174, 212, 236, 180, 176, 228, 196, 144, 204, 210, 120, 180, 222, 120, 168, 226, 144, 152, 226, 180, 160, 214, 228, 160, 162, 160, 0, 0, 0, 0, 0, 0}

type PPU struct {
	Bus       *Bus
	ImageData []uint8
	Registers PPURegisters

	CycleCount uint64
	FrameCount uint64
	Line       uint64

	Nametables     [4 * 0x400]uint8
	PaletteStorage [0x20]uint8
	OAMData        [256]uint8 // 64 entries of 4 bytes: y, tile, attributes, x; in this order
	OAMAddr        uint8
	PPUAddr        uint16
	TempAddr       uint16
	ReadData       uint8

	//
	NMI_Triggered bool
	NMI_Delay     int

	// Tile information
	Tile TileData
}

type TileData struct {
	Id         uint32
	Attributes uint32
	NameTable  uint8

	NextId         uint32
	NextAttributes uint32
}

type PPURegisters struct { // TODO: Update LaTeX file with the write/read restrictions
	PPUCTRL   PPUCTRLRegister   // 0x2000 Write only
	PPUMASK   PPUMASKRegister   // 0x2001 Write only
	PPUSTATUS PPUSTATUSRegister // 0x2002 Read only
	OAMADDR   OAMADDRRegister   // 0x2003 Write only
	OAMDATA   OAMDATARegister   // 0x2004 Read/Write; TODO: Writes increment OAMADDR
	PPUSCROLL PPUSCROLLRegister // 0x2005 Write x2 only
	PPUADDR   PPUADDRRegister   // 0x2006 Write x2 only
	PPUDATA   PPUDATARegister   // 0x2007 Read/Write
	OAMDMA    OAMDMARegister    // 0x4014 Write only

	PPUSCROLL_Y                  bool // The next write to PPUSCROLL will be the Y scroll value
	PPUADDR_LeastSignificantByte bool // The next write to PPUADDR will be the least significant byte of the address
}

type PPUCTRLRegister struct {
	NametableBase              uint16
	VRAMIncrement              bool
	SpritePatternTableBase     uint16 // Ignored in 8x16 mode
	BackgroundPatternTableBase uint16
	SpriteSize                 bool // false -> 8x8; true -> 8x16
	EXTPins                    bool
	VBlankNMIEnabled           bool

	IgnoreWritesCounter int
}

type PPUMASKRegister struct {
	Greyscale          bool
	ShowLeftBackground bool
	ShowLeftSprites    bool
	ShowBackground     bool
	ShowSprites        bool
	EmphasizeRed       bool
	EmphasizeGreen     bool
	EmphasizeBlue      bool
}

type PPUSTATUSRegister struct {
	SpriteOverflow bool
	SpriteZeroHit  bool
	VBlank         bool
}

type OAMADDRRegister struct {
}

type OAMDATARegister struct {
}

type PPUSCROLLRegister struct {
	X uint8
	Y uint8
}

type PPUADDRRegister struct {
}

type PPUDATARegister struct {
}

type OAMDMARegister struct {
}

func (ppu *PPU) Initialize() {
	ppu.ImageData = make([]uint8, 256*240)
	ppu.Registers.PPUCTRL.IgnoreWritesCounter = 30_000
	ppu.Registers.PPUADDR_LeastSignificantByte = false
	ppu.Registers.PPUCTRL.VBlankNMIEnabled = true
	ppu.Registers.PPUSCROLL_Y = false
	ppu.CycleCount = 340
	ppu.FrameCount = 0
	ppu.Line = 240
}

func (ppu *PPU) incementPPUAddr() {
	if ppu.Registers.PPUCTRL.VRAMIncrement {
		ppu.PPUAddr += 32
	} else {
		ppu.PPUAddr++
	}
}

func (ppu *PPU) ReadRegister(address uint16) uint8 {
	switch address {
	case 0x2000:
		panic("PPUCTRL register is write only")
	case 0x2001:
		panic("PPUMASK register is write only")
	case 0x2002:
		var value uint8 = 0
		if ppu.Registers.PPUSTATUS.SpriteOverflow {
			value |= 1 << 5
		}
		if ppu.Registers.PPUSTATUS.SpriteZeroHit {
			value |= 1 << 6
		}
		if ppu.Registers.PPUSTATUS.VBlank {
			value |= 1 << 7
		}

		ppu.Registers.PPUSTATUS.VBlank = false
		if ppu.NMI_Triggered && ppu.NMI_Delay == 0 && ppu.Registers.PPUCTRL.VBlankNMIEnabled {
			ppu.NMI_Delay = 15
		}
		return value
	case 0x2003:
		panic("OAMADDR register is write only")
	case 0x2004:
		return ppu.OAMData[ppu.OAMAddr]
	case 0x2005:
		panic("PPUSCROLL register is write only")
	case 0x2006:
		panic("PPUADDR register is write only")
	case 0x2007:
		value := ppu.Read(ppu.PPUAddr)
		ppu.incementPPUAddr()
		return value
	case 0x4014:
		panic("OAMDMA register is write only")
	}
	return 0
}

func (ppu *PPU) Read(address uint16) uint8 {
	switch {
	case address < 0x2000: // pattern tables, on the cartridge
		ppu.ReadData = ppu.Bus.nes.Cartridge.CHR_ROM[address]
		return ppu.ReadData
	case address < 0x3F00: // name tables
		ppu.ReadData = ppu.Nametables[(address-0x2000)%0x1000]
		return ppu.ReadData
	case address < 0x4000: // palette
		ppu.ReadData = ppu.PaletteStorage[(address-0x3F00)%0x20]
		return ppu.ReadData
	default:
		panic("Invalid PPU address: " + fmt.Sprintf("%x", address))
	}
}

func (ppu *PPU) WriteRegister(address uint16, value uint8) {
	switch address {
	case 0x2000:
		ppu.Registers.PPUCTRL.NametableBase = uint16(value & 0x3)
		ppu.Registers.PPUCTRL.VRAMIncrement = (value & 0x4) != 0
		if (value & 0x8) != 0 {
			ppu.Registers.PPUCTRL.SpritePatternTableBase = 0x1000
		} else {
			ppu.Registers.PPUCTRL.SpritePatternTableBase = 0x0000
		}
		if (value & 0x10) != 0 {
			ppu.Registers.PPUCTRL.BackgroundPatternTableBase = 0x1000
		} else {
			ppu.Registers.PPUCTRL.BackgroundPatternTableBase = 0x0000
		}
		ppu.Registers.PPUCTRL.SpriteSize = (value & 0x20) != 0
		ppu.Registers.PPUCTRL.EXTPins = (value & 0x40) != 0
		ppu.Registers.PPUCTRL.VBlankNMIEnabled = (value & 0x80) != 0
		if ppu.NMI_Triggered && ppu.NMI_Delay == 0 && ppu.Registers.PPUCTRL.VBlankNMIEnabled {
			ppu.NMI_Delay = 15
		}
		ppu.TempAddr = (ppu.TempAddr & 0xF3FF) | ((uint16(value) & 0x3) << 10)
		// TODO: NMI
	case 0x2001:
		ppu.Registers.PPUMASK.Greyscale = (value & 0x1) != 0
		ppu.Registers.PPUMASK.ShowLeftBackground = (value & 0x2) != 0
		ppu.Registers.PPUMASK.ShowLeftSprites = (value & 0x4) != 0
		ppu.Registers.PPUMASK.ShowBackground = (value & 0x8) != 0
		ppu.Registers.PPUMASK.ShowSprites = (value & 0x10) != 0
		ppu.Registers.PPUMASK.EmphasizeRed = (value & 0x20) != 0
		ppu.Registers.PPUMASK.EmphasizeGreen = (value & 0x40) != 0
		ppu.Registers.PPUMASK.EmphasizeBlue = (value & 0x80) != 0
	case 0x2002:
		panic("Not writable")
	case 0x2003:
		ppu.OAMAddr = value
	case 0x2004:
		ppu.OAMData[ppu.OAMAddr] = value
		ppu.OAMAddr++
	case 0x2005:
		if ppu.Registers.PPUSCROLL_Y {
			ppu.TempAddr = (ppu.TempAddr & 0x8FFF) | ((uint16(value) & 0x07) << 12)
			ppu.TempAddr = (ppu.TempAddr & 0xFC1F) | ((uint16(value) & 0xF8) << 2)
		} else {
			ppu.TempAddr = (ppu.TempAddr & 0xFFE0) | (uint16(value) >> 3)
			ppu.Registers.PPUSCROLL.X = value & 0x07
		}
		ppu.Registers.PPUSCROLL_Y = !ppu.Registers.PPUSCROLL_Y
	case 0x2006:
		if ppu.Registers.PPUADDR_LeastSignificantByte {
			ppu.TempAddr = (ppu.TempAddr & 0xFF00) | uint16(value)
			ppu.PPUAddr = ppu.TempAddr
		} else {
			ppu.TempAddr = (ppu.TempAddr & 0x80FF) | (uint16(value) << 8)
		}
		ppu.Registers.PPUADDR_LeastSignificantByte = !ppu.Registers.PPUADDR_LeastSignificantByte
	case 0x2007:
		ppu.Write(ppu.PPUAddr, value)
		ppu.incementPPUAddr()
	case 0x4014:
		page := value
		starting_address := uint16(page) << 8

		for i := 0; i < 256; i++ {
			ppu.OAMData[i] = ppu.Bus.Read(starting_address + uint16(i))
		}

		ppu.Bus.nes.CPU.CycleCount += 513
		if ppu.Bus.nes.CPU.CycleCount%2 == 1 {
			ppu.Bus.nes.CPU.CycleCount++
		}
		ppu.CycleCount += 171
	default:
		panic("Invalid PPU register")
	}
}

func (ppu *PPU) Write(address uint16, value uint8) {
	switch {
	case address < 0x2000: // pattern tables, on the cartridge
		ppu.Bus.nes.Cartridge.Write(address, value)
	case address < 0x3F00: // name tables
		//TODO: Mirroring to be implemented
		ppu.Nametables[ppu.Registers.PPUCTRL.NametableBase+(address-0x2000)%0x1000] = value
	case address < 0x4000: // palette
		ppu.PaletteStorage[(address-0x3F00)%0x20] = value
	}
}

func (ppu *PPU) vBlank() {
	ppu.NMI_Triggered = true
	if ppu.NMI_Triggered && ppu.NMI_Delay == 0 && ppu.Registers.PPUCTRL.VBlankNMIEnabled {
		ppu.NMI_Delay = 15
	}
}

func (ppu *PPU) getBackgroundPixel() uint8 {
	if ppu.Registers.PPUMASK.ShowBackground {
		return uint8(ppu.Tile.Attributes >> ((7 - (ppu.Registers.PPUSCROLL.X)) * 2))
	} else {
		return 0x00
	}
}

func (ppu *PPU) renderPixel() {
	x := ppu.CycleCount - 1
	y := ppu.Line

	background := ppu.getBackgroundPixel()

	if x < 8 && ppu.Registers.PPUMASK.ShowLeftBackground {
		background = 0x00
	}

	color := background

	ppu.ImageData[x+y*256] = ppu.Read(0x3F00 + uint16(color))
}

func (ppu *PPU) Cycle() {
	if ppu.NMI_Delay > 0 {
		ppu.NMI_Delay--
		if ppu.NMI_Delay == 0 && ppu.Registers.PPUCTRL.VBlankNMIEnabled && ppu.NMI_Triggered {
			ppu.Registers.PPUCTRL.VBlankNMIEnabled = false
			ppu.Bus.VBlank()
		}
	}

	// If rendering is enabled (of any type)
	if ppu.Registers.PPUMASK.ShowBackground || ppu.Registers.PPUMASK.ShowSprites {
		// Visible pixel
		if ppu.Line < 240 && ppu.CycleCount >= 1 && ppu.CycleCount <= 256 {
			ppu.renderPixel()
		}

		// Line and cycle for fetching tile information
		if (ppu.Line < 240 || ppu.Line == 261) && ((ppu.CycleCount >= 321 && ppu.CycleCount <= 336) || (ppu.CycleCount >= 1 && ppu.CycleCount <= 256)) {
			ppu.Tile.Id <<= 2
			ppu.Tile.Attributes <<= 2

			switch ppu.CycleCount & 0x07 {
			case 0:
				ppu.Tile.Id |= uint32(ppu.Tile.NextId & 0x03)
				ppu.Tile.Attributes |= uint32(ppu.Tile.NextAttributes)
			case 1:
				ppu.Tile.NameTable = ppu.Read(0x2000 | (ppu.PPUAddr & 0x0FFF))
			case 3:
				v := ppu.PPUAddr
				address := 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)
				shift := ((v >> 4) & 4) | (v & 2)
				ppu.Tile.NextAttributes = uint32((ppu.Read(address)>>shift)&3) << 2
			case 5:
				fineY := (ppu.PPUAddr >> 12) & 7
				table := ppu.Registers.PPUCTRL.BackgroundPatternTableBase
				tile := ppu.Tile.NameTable
				address := table + uint16(tile)*16 + fineY
				value := ppu.Read(address)
				var to_store uint32 = 0
				for i := 0; i < 8; i++ {
					to_store |= ((uint32(value) >> (7 - i)) & 0x1) << ((7 - i) * 2)
				}
				ppu.Tile.NextId = to_store
			case 7:
				fineY := (ppu.PPUAddr >> 12) & 7
				table := ppu.Registers.PPUCTRL.BackgroundPatternTableBase
				tile := ppu.Tile.NameTable
				address := table + uint16(tile)*16 + fineY
				value := ppu.Read(address + 8)
				var to_store uint32 = 0
				for i := 0; i < 8; i++ {
					to_store |= ((uint32(value) >> (7 - i)) & 0x1) << ((7-i)*2 + 1)
				}
				ppu.Tile.NextId |= to_store
			}
		}

		if ppu.Line < 240 || ppu.Line == 261 {
			if (ppu.CycleCount >= 321 && ppu.CycleCount <= 336) || (ppu.CycleCount >= 1 && ppu.CycleCount <= 256) {
				v := ppu.PPUAddr
				if v&0x1F == 31 {
					v &= 0xFFE0
					v ^= 0x0400
				} else {
					v++
				}
				ppu.PPUAddr = v
			}
			if ppu.CycleCount == 256 {
				v := ppu.PPUAddr
				if v&0x7000 != 0x7000 {
					// increment fine Y
					v += 0x1000
				} else {
					// fine Y = 0
					v &= 0x8FFF
					// let y = coarse Y
					y := (v & 0x03E0) >> 5
					if y == 29 {
						// coarse Y = 0
						y = 0
						// switch vertical nametable
						v ^= 0x0800
					} else if y == 31 {
						// coarse Y = 0, nametable not switched
						y = 0
					} else {
						// increment coarse Y
						y++
					}
					// put coarse Y back into v
					v = (v & 0xFC1F) | (y << 5)
				}
				ppu.PPUAddr = v
			}
			if ppu.CycleCount == 257 {
				ppu.PPUAddr = (ppu.PPUAddr & 0x841F) | (ppu.TempAddr & 0x7BE0)
			}
		}

	}

	// Move to the next pixel
	ppu.CycleCount++
	if ppu.CycleCount > 340 {
		ppu.CycleCount = 0
		ppu.Line++
		if ppu.Line > 261 {
			ppu.Line = 0
			ppu.FrameCount++
		}
	}

	if ppu.CycleCount == 1 && ppu.Line == 241 {
		ppu.vBlank()
	}

	if ppu.Line == 261 && ppu.CycleCount == 1 {
		ppu.Registers.PPUSTATUS.VBlank = false
		ppu.Registers.PPUSTATUS.SpriteZeroHit = false
		// ppu.Registers.PPUSTATUS.SpriteOverflow = false; Buggy behaviour on original hardware
		if ppu.NMI_Triggered && ppu.NMI_Delay == 0 && ppu.Registers.PPUCTRL.VBlankNMIEnabled {
			ppu.NMI_Delay = 15
		}
	}
}
