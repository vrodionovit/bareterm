package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type TermGrid struct {
	window    *glfw.Window
	program   uint32
	vao       uint32
	vbo       uint32
	cells     [][]rune
	cellSize  [2]float32
	textColor [4]float32
	bgColor   [4]float32
	font      *Font
	cursor    [2]int
}

func NewTermGrid(width, height int, rows, cols int) (*TermGrid, error) {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "TermGrid", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %v", err)
	}

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenGL: %v", err)
	}

	grid := &TermGrid{
		window:    window,
		cells:     make([][]rune, rows),
		cellSize:  [2]float32{float32(width) / float32(cols), float32(height) / float32(rows)},
		textColor: [4]float32{1, 1, 1, 1},
		bgColor:   [4]float32{0, 0, 0, 1},
		cursor:    [2]int{0, 0},
	}
	for i := range grid.cells {
		grid.cells[i] = make([]rune, cols)
	}

	if err := grid.initOpenGL(); err != nil {
		return nil, err
	}

	grid.font, err = NewFont("DejaVuSansMono.ttf", int(grid.cellSize[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to create font: %v", err)
	}

	return grid, nil
}

func (g *TermGrid) initOpenGL() error {
	var err error
	g.program, err = createShaderProgram()
	if err != nil {
		return fmt.Errorf("failed to create shader program: %v", err)
	}

	gl.GenVertexArrays(1, &g.vao)
	gl.BindVertexArray(g.vao)

	gl.GenBuffers(1, &g.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)

	vertices := []float32{
		0, 0, 0, 0,
		1, 0, 1, 0,
		0, 1, 0, 1,
		1, 1, 1, 1,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	return nil
}

func (g *TermGrid) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(g.program)

	gl.Uniform2fv(gl.GetUniformLocation(g.program, gl.Str("cellSize\x00")), 1, &g.cellSize[0])
	gl.Uniform4fv(gl.GetUniformLocation(g.program, gl.Str("textColor\x00")), 1, &g.textColor[0])
	gl.Uniform4fv(gl.GetUniformLocation(g.program, gl.Str("bgColor\x00")), 1, &g.bgColor[0])

	for row, line := range g.cells {
		for col, char := range line {
			if char == 0 {
				continue
			}
			g.renderCell(row, col, char)
		}
	}

	g.window.SwapBuffers()
}

func (g *TermGrid) renderCell(row, col int, char rune) {
	width, height := g.window.GetSize()
	position := [2]float32{
		2*float32(col)*g.cellSize[0]/float32(width) - 1,
		1 - 2*float32(row+1)*g.cellSize[1]/float32(height),
	}
	gl.Uniform2fv(gl.GetUniformLocation(g.program, gl.Str("cellPosition\x00")), 1, &position[0])

	texture := g.font.GetCharTexture(char)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func (g *TermGrid) SetCell(row, col int, char rune) {
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[0]) {
		g.cells[row][col] = char
	}
}

func (g *TermGrid) SetText(text string) {
	lines := strings.Split(text, "\n")
	for row, line := range lines {
		if row >= len(g.cells) {
			break
		}
		for col, char := range line {
			if col >= len(g.cells[row]) {
				break
			}
			g.cells[row][col] = char
		}
	}
	g.cursor[0] = len(lines) - 1
	g.cursor[1] = len(lines[len(lines)-1])
}

func (g *TermGrid) AppendChar(char rune) {
	if g.cursor[1] < len(g.cells[g.cursor[0]]) {
		g.cells[g.cursor[0]][g.cursor[1]] = char
		g.cursor[1]++
	}
}

func (g *TermGrid) NewLine() {
	if g.cursor[0] < len(g.cells)-1 {
		g.cursor[0]++
		g.cursor[1] = 0
	}
}

func (g *TermGrid) Backspace() {
	if g.cursor[1] > 0 {
		g.cursor[1]--
		g.cells[g.cursor[0]][g.cursor[1]] = 0
	} else if g.cursor[0] > 0 {
		g.cursor[0]--
		g.cursor[1] = len(g.cells[g.cursor[0]]) - 1
		for g.cursor[1] > 0 && g.cells[g.cursor[0]][g.cursor[1]-1] == 0 {
			g.cursor[1]--
		}
	}
}

func (g *TermGrid) Destroy() {
	gl.DeleteProgram(g.program)
	gl.DeleteVertexArrays(1, &g.vao)
	gl.DeleteBuffers(1, &g.vbo)
	g.font.Destroy()
}
