package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	grid, err := NewTermGrid(800, 600, 20, 40)
	if err != nil {
		log.Fatalln("failed to create TermGrid:", err)
	}
	defer grid.Destroy()

	grid.SetText("Hello, TermGrid!")

	grid.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			if key == glfw.KeyEscape {
				w.SetShouldClose(true)
			} else if key >= glfw.KeySpace && key <= glfw.KeyGraveAccent {
				grid.AppendChar(rune(key))
			} else if key == glfw.KeyEnter {
				grid.NewLine()
			} else if key == glfw.KeyBackspace {
				grid.Backspace()
			}
		}
	})

	for !grid.window.ShouldClose() {
		grid.Render()
		glfw.PollEvents()
	}

	fmt.Println("Exiting...")
}
