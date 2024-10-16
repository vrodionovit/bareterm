package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// Эта функция гарантирует, что main() будет выполняться в главном потоке.
	// Это необходимо для GLFW, так как некоторые его функции должны вызываться из главного потока.
	runtime.LockOSThread()
}

func main() {
	// Инициализация GLFW. Это необходимо сделать перед использованием любых функций GLFW.
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	// Отложенный вызов Terminate() гарантирует, что GLFW будет корректно завершен при выходе из программы.
	defer glfw.Terminate()

	// Создание окна с явным указанием названия
	window, err := glfw.CreateWindow(800, 600, "Bareterm Terminal Emulator", nil, nil)
	if err != nil {
		log.Fatalln("failed to create window:", err)
	}

	// Создание нового экземпляра TermGrid с заданными размерами окна и сетки.
	grid, err := NewTermGrid(800, 600, 20, 40)
	if err != nil {
		log.Fatalln("failed to create TermGrid:", err)
	}
	defer grid.Destroy()

	grid.SetText("Hello, TermGrid!")

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			if key == glfw.KeyEscape {
				// Закрытие окна при нажатии Escape.
				w.SetShouldClose(true)
			} else if key >= glfw.KeySpace && key <= glfw.KeyGraveAccent {
				// Добавление символа в сетку при нажатии печатаемых клавиш.
				grid.AppendChar(rune(key))
			} else if key == glfw.KeyEnter {
				// Переход на новую строку при нажатии Enter.
				grid.NewLine()
			} else if key == glfw.KeyBackspace {
				// Удаление последнего символа при нажатии Backspace.
				grid.Backspace()
			}
		}
	})

	// Основной цикл приложения с явным рендерингом
	for !window.ShouldClose() {
		// Очистка экрана
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// Рендеринг TermGrid
		grid.Render()

		// Обмен буферов и обработка событий
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
