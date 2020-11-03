package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"
	"math"
	"github.com/go-gl/gl/v2.1/gl" // OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	width              = 500
	height             = 500
	vertexShaderSource = `
		#version 120
		attribute vec3 vertexPosition_modelspace;
		void main() {
			gl_Position.xyz = vertexPosition_modelspace;
			gl_Position.w = 1.0;
		}
		` + "\x00"

	fragmentShaderSource = `
	#version 120
	void main() {
		gl_FragColor = vec4(0.5, 0.89, 0.74, 1.0);
	}
	` + "\x00"
)

var (
	triangle = Triangle{-1,-1,-0.5,-1,0.5,-0.5}
	rect = Rect{-0.5,0.5,-0.5,-0.5,0.5,-0.5, 0.5,0.5}
	player = Player{Rect{}, Triangle{}, Triangle{}, Triangle{}, 0,0,math.Pi/4}
	//[]float32{
		// 0, 0.5, 0,
		// -0.5, -0.5, 0,
		// 0.5, -0.5, 0,
	// }
)

type Triangle struct{
	X1 float32
	Y1 float32
	X2 float32
	Y2 float32
	X3 float32
	Y3 float32
}

type Player struct{
	body Rect
	hat Triangle
	wing1 Triangle
	wing2 Triangle
	x float32
	y float32
	direction float64
}

func (player *Player) draw(x float32, y float32, scale float32) {
	player.body.X1 = x + scale*float32(math.Cos(player.direction-math.Pi/6))
	player.body.Y1 = y + scale*float32(math.Sin(player.direction-math.Pi/6))
	player.body.X2 = x + scale*float32(math.Cos(player.direction+math.Pi/6))
	player.body.Y2 = y + scale*float32(math.Sin(player.direction+math.Pi/6))
	player.body.X3 = x + scale*float32(math.Cos(player.direction+math.Pi/6+math.Pi))
	player.body.Y3 = y + scale*float32(math.Sin(player.direction+math.Pi/6+math.Pi))
	player.body.X4 = x + scale*float32(math.Cos(player.direction-math.Pi/6+math.Pi))
	player.body.Y4 = y + scale*float32(math.Sin(player.direction-math.Pi/6+math.Pi))
	player.hat.X1 = player.body.X1
	player.hat.Y1 = player.body.Y1
	player.hat.X2 = player.body.X2
	player.hat.Y2 = player.body.Y2
	player.hat.X3 = x + (scale+0.1)*float32(math.Cos(player.direction))
	player.hat.X3 = x + (scale+0.1)*float32(math.Sin(player.direction))
	player.body.draw()
	player.hat.draw()
}


func (a Triangle) getVao() uint32{
	return makeVao(a.getArray())
}
func (a Rect) getVao() uint32{
	return makeVao(a.getArray())
}
func (a Triangle) getArray() ([]float32) {
	return []float32{a.X1, a.Y1, 0, a.X2, a.Y2, 0, a.X3, a.Y3,0}
}
func (c *Rect) draw() {
    gl.BindVertexArray(c.getVao())
    gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.getArray()) / 3))
}
func (c *Triangle) draw() {
    gl.BindVertexArray(c.getVao())
    gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.getArray()) / 3))
}
type Rect struct{
	X1 float32
	Y1 float32
	X2 float32
	Y2 float32
	X3 float32
	Y3 float32
	X4 float32
	Y4 float32
}

func (a Rect) getArray() ([]float32) {
	return []float32{
	a.X1, a.Y1, 0,
	a.X2, a.Y2, 0,
	a.X3, a.Y3, 0,
	a.X3, a.Y3, 0,
	a.X2, a.Y2, 0,
	a.X4, a.Y4, 0}
}
func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	program := initOpenGL()
	for !window.ShouldClose() {
		time.Sleep(time.Duration(time.Second/24))
		update()
		draw(window, program)
	}
}

func draw(window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)
	// rect.draw()
	// triangle.draw()
	player.draw(0,0,0.2)
	// gl.BindVertexArray(vao)
	// gl.DrawArrays(gl.TRIANGLES, 0, int32(len(rect.getArray())/3))
	glfw.PollEvents()
	window.SwapBuffers()
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
		glfw.WindowHint(glfw.Resizable, glfw.True)
	     	// glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	 //	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// initOpenGL initializes OpenGL and returns an intiialized program.
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

// makeVao initializes and returns a vertex array from the points provided.
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
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
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

func update() {

}
