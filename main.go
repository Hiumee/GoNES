package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hiumee/NES/internals"
)

const (
	FREQUENCY = 1789773
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
		float col = texture(gameTexture, texCoo).r*4.0;
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
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 256, 240, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(buffer))
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
var Config = flag.String("config", "", "Configuration file for the emulator containing the keyboard mapping")

var cpuprofile = ""

var USER_INPUT struct {
	A, B, Select, Start, Up, Down, Left, Right, Reset glfw.Key
}

type ConfigS struct {
	Keys ConfigKeys `json:"keys"`
}

type ConfigKeys struct {
	Up     string `json:"up"`
	Down   string `json:"down"`
	Left   string `json:"left"`
	Right  string `json:"right"`
	A      string `json:"a"`
	B      string `json:"b"`
	Select string `json:"select"`
	Start  string `json:"start"`
	Reset  string `json:"reset"`
}

func getKeyCode(key string) (glfw.Key, error) {
	character := key[0]
	if character >= 'a' && character <= 'z' {
		character = character - 'a' + 'A'
	}
	if character >= 'A' && character <= 'Z' {
		return glfw.Key(character), nil
	} else {
		return glfw.Key(0), fmt.Errorf("invalid key %s", key)
	}
}

func loadConfig() {
	USER_INPUT.A = glfw.KeyI
	USER_INPUT.B = glfw.KeyO
	USER_INPUT.Select = glfw.KeyK
	USER_INPUT.Start = glfw.KeyL
	USER_INPUT.Up = glfw.KeyW
	USER_INPUT.Down = glfw.KeyS
	USER_INPUT.Left = glfw.KeyA
	USER_INPUT.Right = glfw.KeyD
	USER_INPUT.Reset = glfw.KeyR

	if *Config != "" {
		configData, err := ioutil.ReadFile(*Config)
		if err != nil {
			log.Println("Could not read the configuration file. Using the default configuration")
		}

		var config ConfigS
		json.Unmarshal(configData, &config)

		if config.Keys.A != "" {
			key, err := getKeyCode(config.Keys.A)
			if err != nil {
				log.Println("Invalid key for A:", config.Keys.A)
			} else {
				USER_INPUT.A = key
			}
		}

		if config.Keys.B != "" {
			key, err := getKeyCode(config.Keys.B)
			if err != nil {
				log.Println("Invalid key for B:", config.Keys.B)
			} else {
				USER_INPUT.B = key
			}
		}

		if config.Keys.Select != "" {
			key, err := getKeyCode(config.Keys.Select)
			if err != nil {
				log.Println("Invalid key for Select:", config.Keys.Select)
			} else {
				USER_INPUT.Select = key
			}
		}

		if config.Keys.Start != "" {
			key, err := getKeyCode(config.Keys.Start)
			if err != nil {
				log.Println("Invalid key for Start:", config.Keys.Start)
			} else {
				USER_INPUT.Start = key
			}
		}

		if config.Keys.Up != "" {
			key, err := getKeyCode(config.Keys.Up)
			if err != nil {
				log.Println("Invalid key for Up:", config.Keys.Up)
			} else {
				USER_INPUT.Up = key
			}
		}

		if config.Keys.Down != "" {
			key, err := getKeyCode(config.Keys.Down)
			if err != nil {
				log.Println("Invalid key for Down:", config.Keys.Down)
			} else {
				USER_INPUT.Down = key
			}
		}

		if config.Keys.Left != "" {
			key, err := getKeyCode(config.Keys.Left)
			if err != nil {
				log.Println("Invalid key for Left:", config.Keys.Left)
			} else {
				USER_INPUT.Left = key
			}
		}

		if config.Keys.Right != "" {
			key, err := getKeyCode(config.Keys.Right)
			if err != nil {
				log.Println("Invalid key for Right:", config.Keys.Right)
			} else {
				USER_INPUT.Right = key
			}
		}

		if config.Keys.Reset != "" {
			key, err := getKeyCode(config.Keys.Reset)
			if err != nil {
				log.Println("Invalid key for Reset:", config.Keys.Reset)
			} else {
				USER_INPUT.Reset = key
			}
		}
	}
}

// 0,16,27,18
func main() {
	if cpuprofile != "" {
		fmt.Println("PROFILING")
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		//pprof.StartCPUProfile(f)
		pprof.WriteHeapProfile(f)
		defer pprof.StopCPUProfile()
	}
	flag.StringVar(ROMFile, "f", "", "alias for `file`")
	flag.BoolVar(PPUViewer, "p", false, "alias for `ppu`")
	flag.StringVar(Palette, "l", "00,12,24,2A", "alias for `palette`")
	flag.StringVar(Config, "c", "", "alias for `config`")

	flag.Parse()

	if *ROMFile == "" {
		flag.Usage()
		return
	}

	loadConfig()

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
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_1D, color_palette_texture)

	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)

	gl.TexImage1D(gl.TEXTURE_1D, 0, gl.RGB, 64, 0, gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(internals.COLOR_PALETTE))

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE1)
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

	gl.UseProgram(program)

	gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("palette\x00")), 0)
	gl.Uniform1i(gl.GetUniformLocation(program, gl.Str("gameTexture\x00")), 1)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.R8, 256, 240, 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(image_data))

	if *PPUViewer {
		for !window.ShouldClose() {
			draw(vao, window, program, image_data)
			glfw.PollEvents()
			time.Sleep(time.Millisecond * 50)
		}
	}

	if !*PPUViewer {
		// Main loop
		start := time.Now()
		ts := start
		for !window.ShouldClose() {
			if nes.Cartridge.Loaded {
				ts = time.Now()
				elapsed := ts.Sub(start).Seconds()
				cycles := int(elapsed * FREQUENCY)
				start = ts
				for cycles > 0 {
					cycles--
					nes.Step()
					if nes.PPU.Line == 241 && (nes.PPU.CycleCount >= 1 && nes.PPU.CycleCount <= 3) {
						for i := 0; i < 256*240; i++ {
							image_data[i] = nes.PPU.ImageData[i]
						}
						draw(vao, window, program, image_data)
						glfw.PollEvents()
						if window.GetKey(USER_INPUT.Reset) == 1 { // A
							nes.CPU.Reset()
						}
						nes.Controllers[0].SetInput(getInput(window))
					}
				}
			}
		}
	}
}

func getInput(window *glfw.Window) [8]bool {
	var input [8]bool
	if window.GetKey(USER_INPUT.A) == 1 { // A
		input[0] = true
	}
	if window.GetKey(USER_INPUT.B) == 1 { // B
		input[1] = true
	}
	if window.GetKey(USER_INPUT.Select) == 1 { // SELECT
		input[2] = true
	}
	if window.GetKey(USER_INPUT.Start) == 1 { // START
		input[3] = true
	}
	if window.GetKey(USER_INPUT.Up) == 1 { // UP
		input[4] = true
	}
	if window.GetKey(USER_INPUT.Down) == 1 { // DOWN
		input[5] = true
	}
	if window.GetKey(USER_INPUT.Left) == 1 { // LEFT
		input[6] = true
	}
	if window.GetKey(USER_INPUT.Right) == 1 { // RIGHT
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
