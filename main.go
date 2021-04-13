package main

import (
	"fmt"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"runtime"
	"time"
	"unsafe"
)

func init() {
	runtime.LockOSThread()
}

const (
	W   = 500
	H   = 500
	fps = 30
)

var (
	viewMat mgl32.Mat4
	projMat mgl32.Mat4
	program uint32
)

func main() {
	// GLFW Initialization
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	// Close glfw when main exits
	defer glfw.Terminate()
	// Window Properties
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	// Create the window with the above hints
	window, err := glfw.CreateWindow(W, H, "Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.SetKeyCallback(HandleKeys)
	window.SetCursorPosCallback(HandleMouseMovement)
	// Load the icon file
	icoFile, err := os.Open("ico.png")
	if err != nil {
		panic(err)
	}
	// decode the file to an image.Image
	ico, err := png.Decode(icoFile)
	if err != nil {
		panic(err)
	}
	fmt.Println(ico.ColorModel())
	window.SetIcon([]image.Image{ico})
	window.MakeContextCurrent()
	// OpenGL Initialization
	// Check for the version
	//version := gl.GoStr(gl.GetString(gl.VERSION))
	//	fmt.Println("OpenGL Version", version)
	// Read the vertex and fragment shader files
	vertexShader, err := ioutil.ReadFile("vertex.vert")
	if err != nil {
		panic(err)
	}
	vertexShader = append(vertexShader, []byte("\x00")...)
	fragmentShader, err := ioutil.ReadFile("frag.frag")
	if err != nil {
		panic(err)
	}
	fragmentShader = append(fragmentShader, []byte("\x00")...)

	err = gl.Init()
	if err != nil {
		panic(err)
	}
	// Set the function for handling errors
	gl.DebugMessageCallback(func(source, gltype, id, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		fmt.Println(source, gltype, severity, id, length, message, userParam)
	}, nil)
	// Create an OpenGL "Program" and link it for current drawing
	program, err = newProg(string(vertexShader), string(fragmentShader))
	if err != nil {
		panic(err)
	}
	// Check for the version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL Version", version)
	// Main draw loop
	// Draw a Shape
	shape := NewShape(mgl32.Ident4(), program)
	shape.Pts = append(shape.Pts,
	PC(0.5, 1, 1, 1, 1, 1, 1),
	PC(1, 1, 1, 1, 1, 1, 1),
	PC(0, 1, 1, 1, 1, 1, 1),
	PC(0, 0, 0, 1, 1, 1, 1),
	PC(0,0,1,1,1,0,1),
)
	// Generate the Vao for the shape
	shape.GenVao()
	shape.SetTypes(gl.TRIANGLE_STRIP)
	// Set the refresh function for the window
	// Use this program
	gl.UseProgram(program)
	// Calculate the projection matrix
	projMat = mgl32.Perspective(mgl32.DegToRad(60), float32(W)/H, 0.1, 10)
	// set the value of Projection matrix
	UpdateUniformMat4fv("projection", program, &projMat[0])
	// Set the value of view matrix
	viewMat = mgl32.Ident4()
	UpdateUniformMat4fv("view", program, &projMat[0])
	for !window.ShouldClose() {
		time.Sleep(time.Second / fps)
		// Clear everything that was drawn previously
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// Actually draw something
		shape.Draw()
		// display everything that was drawn
		window.SwapBuffers()
		// check for any events
		glfw.PollEvents()
	}
}
