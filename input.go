package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	//	"fmt"
	"github.com/go-gl/mathgl/mgl32"
)

func HandleKeys(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	case glfw.KeyE:
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		if glfw.RawMouseMotionSupported() {
			w.SetInputMode(glfw.RawMouseMotion, glfw.True)
		}

	}
}

func HandleMouseMovement(w *glfw.Window, xpos, ypos float64) {
	viewMat = mgl32.LookAt(
		0, 0, 0,
		float32(xpos), float32(ypos), 1,
		0, 1.0, 0,
	)
}
