package internals

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
	ReadData       uint8
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
	OAMDMA    OAMDMARegister    // 0x4014 Read/Write

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
	ppu.Registers.PPUSCROLL_Y = false
	ppu.CycleCount = 340
	ppu.FrameCount = 0
	ppu.Line = 240
}

func (ppu *PPU) IncementPPUAddr() {
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
		switch {
		case ppu.PPUAddr < 0x2000: // pattern tables, on the cartridge
			ppu.ReadData = ppu.Bus.nes.Cartridge.CHR_ROM[ppu.PPUAddr]
			return ppu.ReadData
		case ppu.PPUAddr < 0x3F00: // name tables
			ppu.ReadData = ppu.Nametables[(ppu.PPUAddr-0x2000)%0x1000]
			return ppu.ReadData
		case ppu.PPUAddr < 0x4000: // palette
			ppu.ReadData = ppu.PaletteStorage[(ppu.PPUAddr-0x3F00)%0x20]
			return ppu.ReadData
		}
		ppu.IncementPPUAddr()
	case 0x4014:
		panic("OAMDMA register is write only")
	}
	return 0
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
			ppu.Registers.PPUSCROLL.Y = uint8(value)
		} else {
			ppu.Registers.PPUSCROLL.X = uint8(value)
		}
		ppu.Registers.PPUSCROLL_Y = !ppu.Registers.PPUSCROLL_Y
	case 0x2006:
		if ppu.Registers.PPUADDR_LeastSignificantByte {
			ppu.PPUAddr = (ppu.PPUAddr & 0xFF00) | uint16(value)
		} else {
			ppu.PPUAddr = (ppu.PPUAddr & 0x00FF) | (uint16(value) << 8)
		}
		ppu.Registers.PPUADDR_LeastSignificantByte = !ppu.Registers.PPUADDR_LeastSignificantByte
	case 0x2007:
		switch {
		case ppu.PPUAddr < 0x2000: // pattern tables, on the cartridge
			ppu.Bus.nes.Cartridge.Write(ppu.PPUAddr, value)
		case ppu.PPUAddr < 0x3F00: // name tables
			//TODO: Mirroring to be implemented
			ppu.Nametables[ppu.Registers.PPUCTRL.NametableBase+(ppu.PPUAddr-0x2000)%0x1000] = value
		case ppu.PPUAddr < 0x4000: // palette
			ppu.PaletteStorage[(ppu.PPUAddr-0x3F00)%0x20] = value
		}
		ppu.IncementPPUAddr()
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

func (ppu *PPU) VBlank() {
	panic("Not implemented")
}

func (ppu *PPU) Cycle() {
	panic("Not implemented")
}
