package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hiumee/NES/internals"
)

const (
	vertexShaderSource = `
    #version 410
    in vec2 vp;
	out vec2 texCoo;
    void main() {
        gl_Position = vec4(vp.x, vp.y, 0.0, 1.0);
		texCoo = (vec2(vp.x, -vp.y)+1)/2;
    }
` + "\x00"

	fragmentShaderSource = `
    #version 410
	in vec2 texCoo;
    out vec4 frag_colour;
	uniform sampler2D gameTexture;
	uniform sampler1D palette;
    void main() {
		float col = texture(gameTexture, texCoo).r*4.0f;
		frag_colour = texture(palette, col);
    }
` + "\x00"
)

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(256*4, 240*4, "NES", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

var (
	triangle = []float32{
		-1, 1, // top left
		-1, -1, // bottom left
		1, 1, // top right
		1, -1, // bottom right
	}
)

func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*2, nil)

	return vao
}

func draw(vao uint32, window *glfw.Window, program uint32, buffer []uint8) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(program)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R8, 256, 240, 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(buffer))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(len(triangle)/2))

	window.SwapBuffers()
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

var image_data []uint8 = make([]uint8, 256*240)

var ROMFile = flag.String("file", "", "ROM file to load")
var PPUViewer = flag.Bool("ppu", false, "Show PPU viewer")
var Palette = flag.String("palette", "00,12,24,2A", "Palette information to use. Must be 4 hexadecimal representation of colors separated by commas (0x00-0x3F)")

// 0,16,27,18
func main() {
	flag.StringVar(ROMFile, "f", "", "alias for `file`")
	flag.BoolVar(PPUViewer, "p", false, "alias for `ppu`")
	flag.StringVar(Palette, "l", "00,12,24,2A", "alias for `palette`")

	flag.Parse()

	if *ROMFile == "" {
		flag.Usage()
		return
	}

	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	program := initOpenGL()

	hexColors := strings.Split(*Palette, ",")
	if len(hexColors) != 4 {
		log.Fatal("Invalid palette. Must be 4 hexadecimal representation of colors separated by commas (0x00-0x3F). Default: \"00,12,24,2A\". Provided value: ", *Palette)
	}
	var color_palette []uint8 = make([]uint8, 4)
	for i, hexColor := range hexColors {
		color, _ := strconv.ParseUint(hexColor, 16, 8)
		color_palette[i] = uint8(color)
	}

	var color_palette_texture uint32

	gl.GenTextures(1, &color_palette_texture)
	gl.BindTexture(gl.TEXTURE_1D, color_palette_texture)

	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)

	gl.TexImage1D(gl.TEXTURE_1D, 0, gl.RGB, 64, 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(internals.COLOR_PALETTE))

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	vao := makeVao(triangle)
	gl.BindVertexArray(vao)

	nes := internals.NewNES()
	nes.LoadFile(*ROMFile)

	patterns := nes.Cartridge.CHR_ROM

	line := -1
	for k := 0; k < int(nes.Cartridge.Header.CHR_ROM_size)/16; k++ {
		pattern := getPattern(patterns, uint(k))

		if k%32 == 0 {
			line++
		}
		for i := 0; i < 8; i++ {
			for j := 0; j < 8; j++ {
				color := pattern[i*8+j]

				image_data[(i+line*8)*256+j+8*(k%32)] = color_palette[color]
			}
		}
	}

	// Main loop
	//start := time.Now()
	for !window.ShouldClose() {
		if nes.Cartridge.Loaded {
			nes.Step()
			if nes.PPU.Line == 241 && (nes.PPU.CycleCount == 1 || nes.PPU.CycleCount == 2 || nes.PPU.CycleCount == 3) {
				for i := 0; i < 256*240; i++ {
					image_data[i] = nes.PPU.ImageData[i]
				}
				draw(vao, window, program, image_data)
				glfw.PollEvents()
				nes.Controllers[0].SetInput(getInput(window))
				//elapsed := time.Since(start)
				//time.Sleep((16 - time.Duration(elapsed.Milliseconds())) * time.Millisecond)
				//start = time.Now()
			}
		}
	}
}

func getInput(window *glfw.Window) [8]bool {
	var input [8]bool
	if window.GetKey(glfw.KeyO) == 1 { // A
		input[0] = true
	}
	if window.GetKey(glfw.KeyI) == 1 { // B
		input[1] = true
	}
	if window.GetKey(glfw.KeyL) == 1 { // SELECT
		input[2] = true
	}
	if window.GetKey(glfw.KeyK) == 1 { // START
		input[3] = true
	}
	if window.GetKey(glfw.KeyW) == 1 { // UP
		input[4] = true
	}
	if window.GetKey(glfw.KeyS) == 1 { // DOWN
		input[5] = true
	}
	if window.GetKey(glfw.KeyA) == 1 { // LEFT
		input[6] = true
	}
	if window.GetKey(glfw.KeyD) == 1 { // RIGHT
		input[7] = true
	}
	return input
}

func getPattern(patterns []byte, index uint) [64]uint8 {
	pattern := patterns[index*16 : index*16+16]
	var pixels [64]byte

	for i := 0; i < 8; i++ {
		pixels[i*8] = (pattern[i] >> 7) & 1
		pixels[i*8+1] = (pattern[i] >> 6) & 1
		pixels[i*8+2] = (pattern[i] >> 5) & 1
		pixels[i*8+3] = (pattern[i] >> 4) & 1
		pixels[i*8+4] = (pattern[i] >> 3) & 1
		pixels[i*8+5] = (pattern[i] >> 2) & 1
		pixels[i*8+6] = (pattern[i] >> 1) & 1
		pixels[i*8+7] = (pattern[i] >> 0) & 1
	}
	for i := 0; i < 8; i++ {
		pixels[i*8] |= ((pattern[i+8] >> 7) & 1) << 1
		pixels[i*8+1] |= ((pattern[i+8] >> 6) & 1) << 1
		pixels[i*8+2] |= ((pattern[i+8] >> 5) & 1) << 1
		pixels[i*8+3] |= ((pattern[i+8] >> 4) & 1) << 1
		pixels[i*8+4] |= ((pattern[i+8] >> 3) & 1) << 1
		pixels[i*8+5] |= ((pattern[i+8] >> 2) & 1) << 1
		pixels[i*8+6] |= ((pattern[i+8] >> 1) & 1) << 1
		pixels[i*8+7] |= ((pattern[i+8] >> 0) & 1) << 1
	}

	return pixels
}
