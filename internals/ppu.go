package internals

type PPU struct {
	ImageData []uint8
	Registers PPURegisters
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
	OAMAddress uint16
}

type OAMDATARegister struct {
}

type PPUSCROLLRegister struct {
}

type PPUADDRRegister struct {
}

type PPUDATARegister struct {
}

type OAMDMARegister struct {
}

func (ppu *PPU) Initialize() {
	ppu.ImageData = make([]uint8, 256*240*3)
	ppu.Registers.PPUCTRL.IgnoreWritesCounter = 30_000
}

func (ppu *PPU) ReadRegister(address uint16) uint8 {
	panic("Not implemented")
}

func (ppu *PPU) WriteRegister(address uint16, value uint8) {
	panic("Not implemented")
}

func (ppu *PPU) Cycle() {
	panic("Not implemented")
}
