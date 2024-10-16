package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// TermGrid представляет собой структуру для отображения сетки символов.
type TermGrid struct {
	window      *glfw.Window // Окно GLFW для отображения сетки
	program     uint32       // Идентификатор шейдерной программы OpenGL
	vao         uint32       // Vertex Array Object для хранения состояния вершинных атрибутов
	vbo         uint32       // Vertex Buffer Object для хранения вершинных данных
	cells       [][]rune     // Двумерный массив для хранения символов в сетке
	cellSize    [2]float32   // Размер одной ячейки сетки (ширина, высота)
	textColor   [4]float32   // Цвет текста (RGBA)
	bgColor     [4]float32   // Цвет фона (RGBA)
	font        *Font        // Шрифт для отрисовки текста
	cursor      [2]int       // Позиция курсора в сетке (строка, столбец)
	needsRedraw bool         // Флаг необходимости перерисовки
}

// NewTermGrid создает новый экземпляр TermGrid с заданными размерами.
func NewTermGrid(width, height int, rows, cols int) (*TermGrid, error) {
	// Устанавливаем подсказки для создания окна GLFW
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Создаем окно GLFW
	window, err := glfw.CreateWindow(width, height, "TermGrid", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %v", err)
	}

	// Устанавливаем текущий контекст OpenGL
	window.MakeContextCurrent()

	// Инициализируем OpenGL
	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenGL: %v", err)
	}

	// Создаем и инициализируем структуру TermGrid
	grid := &TermGrid{
		window:      window,
		cells:       make([][]rune, rows),
		cellSize:    [2]float32{float32(width) / float32(cols), float32(height) / float32(rows)},
		textColor:   [4]float32{1, 1, 1, 1}, // Белый цвет по умолчанию
		bgColor:     [4]float32{0, 0, 0, 1}, // Черный цвет по умолчанию
		cursor:      [2]int{0, 0},
		needsRedraw: true,
	}
	for i := range grid.cells {
		grid.cells[i] = make([]rune, cols)
	}

	// Инициализируем OpenGL ресурсы
	if err := grid.initOpenGL(); err != nil {
		return nil, err
	}
	// Устанавливаем callback для изменения размера окна
	window.SetSizeCallback(grid.ResizeCallback)

	// Создаем шрифт для отрисовки текста
	grid.font, err = NewFont("DejaVuSansMono", int(grid.cellSize[1]))
	if err != nil {
		return nil, fmt.Errorf("failed to create font: %v", err)
	}

	return grid, nil
}

func (g *TermGrid) SetTextColor(color [4]float32) {
	g.textColor = color
	g.needsRedraw = true
}

func (g *TermGrid) SetFontSize(newSize int) error {
	newFont, err := NewFont("DejaVuSansMono", newSize)
	if err != nil {
		return fmt.Errorf("failed to create new font: %v", err)
	}

	g.font.Destroy()
	g.font = newFont

	// Recalculate cell size based on new font size
	width, height := g.window.GetSize()
	g.cellSize = [2]float32{
		float32(width) / float32(len(g.cells[0])),
		float32(height) / float32(len(g.cells)),
	}

	g.needsRedraw = true
	return nil
}

// SetBackgroundColor устанавливает цвет фона для всей сетки.
func (g *TermGrid) SetBackgroundColor(color [4]float32) {
	g.bgColor = color
	g.needsRedraw = true
}

// // Запуск отрисовки
// func (g *TermGrid) Run() {
// 	for !g.window.ShouldClose() {
// 		if g.needsRedraw {
// 			g.Render()
// 			g.needsRedraw = false
// 		}
// 		glfw.PollEvents()
// 	}
// }

// initOpenGL инициализирует необходимые ресурсы OpenGL.
func (g *TermGrid) initOpenGL() error {
	var err error
	// Создаем шейдерную программу
	g.program, err = createShaderProgram()
	if err != nil {
		return fmt.Errorf("failed to create shader program: %v", err)
	}

	// Создаем и настраиваем VAO и VBO
	gl.GenVertexArrays(1, &g.vao)
	gl.BindVertexArray(g.vao)

	gl.GenBuffers(1, &g.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.vbo)

	// Определяем вершины для отрисовки квадрата (ячейки)
	vertices := []float32{
		0, 0, 0, 0,
		1, 0, 1, 0,
		0, 1, 0, 1,
		1, 1, 1, 1,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Настраиваем атрибуты вершин
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	return nil
}

// Render отрисовывает содержимое сетки.
func (g *TermGrid) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.UseProgram(g.program)
	width, height := g.window.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	// Устанавливаем uniform-переменные для шейдеров
	gl.Uniform2fv(gl.GetUniformLocation(g.program, gl.Str("cellSize\x00")), 1, &g.cellSize[0])
	gl.Uniform4fv(gl.GetUniformLocation(g.program, gl.Str("textColor\x00")), 1, &g.textColor[0])
	gl.Uniform4fv(gl.GetUniformLocation(g.program, gl.Str("bgColor\x00")), 1, &g.bgColor[0])

	// Отрисовываем каждую ячейку сетки
	for row, line := range g.cells {
		for col, char := range line {
			if char == 0 {
				continue // Пропускаем пустые ячейки
			}
			g.renderCell(row, col, char)
		}
	}

	g.window.SwapBuffers()
}

// renderCell отрисовывает отдельную ячейку сетки.
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

// SetCell устанавливает символ в указанной позиции сетки.
func (g *TermGrid) SetCell(row, col int, char rune) {
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[0]) {
		g.cells[row][col] = char
		g.needsRedraw = true

	}
}

func (g *TermGrid) SetCell2(row, col int, char rune) {
	if row >= 0 && row < len(g.cells) && col >= 0 && col < len(g.cells[0]) {
		if g.cells[row][col] != char {
			g.cells[row][col] = char
			g.needsRedraw = true
		}
	}
}

// SetText устанавливает текст в сетку, начиная с текущей позиции курсора.
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
	g.needsRedraw = true

}

// AppendChar добавляет символ в текущую позицию курсора.
func (g *TermGrid) AppendChar(char rune) {
	if g.cursor[1] < len(g.cells[g.cursor[0]]) {
		g.cells[g.cursor[0]][g.cursor[1]] = char
		g.cursor[1]++
		g.needsRedraw = true

	}
}

// NewLine переводит курсор на новую строку.
func (g *TermGrid) NewLine() {
	if g.cursor[0] < len(g.cells)-1 {
		g.cursor[0]++
		g.cursor[1] = 0
		g.needsRedraw = true

	}
}

// Backspace удаляет символ перед курсором.
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
	g.needsRedraw = true

}

// Destroy освобождает ресурсы, занятые TermGrid.
func (g *TermGrid) Destroy() {
	gl.DeleteProgram(g.program)
	gl.DeleteVertexArrays(1, &g.vao)
	gl.DeleteBuffers(1, &g.vbo)
	g.font.Destroy()
}

func (g *TermGrid) ResizeCallback(w *glfw.Window, width int, height int) {
	// Обновляем размер viewport OpenGL
	gl.Viewport(0, 0, int32(width), int32(height))

	// Пересчитываем размер ячейки
	g.cellSize = [2]float32{
		float32(width) / float32(len(g.cells[0])),
		float32(height) / float32(len(g.cells)),
	}

	// Обновляем размер шрифта, если необходимо
	newFontSize := int(g.cellSize[1])
	if g.font.size != newFontSize {
		newFont, err := NewFont("DejaVuSansMono", newFontSize)
		if err == nil {
			g.font.Destroy() // Освобождаем ресурсы старого шрифта
			g.font = newFont
		}
	}

	// Устанавливаем флаг необходимости перерисовки
	g.needsRedraw = true
}
