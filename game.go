package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	// "path/filepath"
	"io/ioutil"

	// "unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	width  = 500
	height = 500
	fps    = 30
)

var (
	program              uint32
	window               *glfw.Window
	player               = *new(Player)
	vertexShaderSource   = readfile("vert.glsl")
	fragmentShaderSource = readfile("frag.glsl")
	troops               = *new([5]Enemy)
)

type Triangle struct {
	X1 float32
	Y1 float32
	X2 float32
	Y2 float32
	X3 float32
	Y3 float32
}
type Circle struct {
	x float32
	y float32
	r float32
}

type Enemy struct {
	x         float32
	y         float32
	size      float32
	direction float64
	body      Circle
}

func (e *Enemy) draw() {
	e.x = e.x + float32(math.Cos(e.direction))*0.01
	e.y = e.y + float32(math.Sin(e.direction))*0.01
	e.body.x = e.x
	e.body.y = e.y
	e.body.r = e.size
	// fmt.Println(e.x)
	e.body.draw()
	if e.x > 2 || e.x < -2 || e.y > 2 || e.y < -2 {
		e.spawn()
	}
}

func (e *Enemy) spawn() {
	e.x = rand.Float32() - 0.5
	e.y = rand.Float32() - 0.5
	fmt.Println(e.x, e.y)
	if e.x < 0 {
		e.x -= 1
	}
	if e.x > 0 {
		e.x += 1
	}
	if e.y < 0 {
		e.y -= 1
	}
	if e.y > 0 {
		e.y += 1
	}
	e.size = map1(rand.Float32(), 0, 1, 0.1, 0.15)
	playerdir := math.Atan2(float64(player.y-e.y), float64(player.x-e.x))
	e.direction = playerdir + (0.5*rand.Float64() - 0.25)
	e.draw()
}

func (z *Circle) draw() {
	cx := z.x
	cy := z.y
	r := z.r
	num_segments := 40
	theta := 2 * 3.1415926 / float64(num_segments)
	c := math.Cos(theta) //precalculate the sine and cosine
	s := math.Sin(theta)
	var t float32

	x := r //we start at angle = 0
	var y float32
	circleslice := []float32{}
	circlecolor := []float32{}
	for ii := 0; ii < num_segments; ii++ {
		circleslice = append(circleslice, cx, cy, 0, 0, 0, 0)
		circleslice = append(circleslice, x+cx, y+cy, 0, 0.8, 0.8, 0.8)

		//apply the rotation matrix
		t = x
		x = float32(c)*x - float32(s)*y
		y = float32(s)*t + float32(c)*y
	}
	gl.BindVertexArray(makeVao(circleslice, circlecolor))
	gl.DrawArrays(gl.LINES, 0, int32(len(circleslice))/3)

}

type Player struct {
	body      Rect
	hat       Triangle
	wing1     Triangle
	wing2     Triangle
	x         float32
	y         float32
	direction float64
}

func (play *Player) draw(scale float32) {
	play.body.X1 = player.x + scale*float32(math.Cos(player.direction-math.Pi/6))
	play.body.Y1 = player.y + scale*float32(math.Sin(player.direction-math.Pi/6))
	play.body.X2 = player.x + scale*float32(math.Cos(player.direction+math.Pi/6))
	play.body.Y2 = player.y + scale*float32(math.Sin(player.direction+math.Pi/6))
	play.body.X3 = player.x + scale*float32(math.Cos(player.direction+math.Pi/6+math.Pi))
	play.body.Y3 = player.y + scale*float32(math.Sin(player.direction+math.Pi/6+math.Pi))
	play.body.X4 = player.x + scale*float32(math.Cos(player.direction-math.Pi/6+math.Pi))
	play.body.Y4 = player.y + scale*float32(math.Sin(player.direction-math.Pi/6+math.Pi))
	play.hat.X1 = player.body.X1
	play.hat.Y1 = player.body.Y1
	play.hat.X2 = player.body.X2
	play.hat.Y2 = player.body.Y2
	play.hat.X3 = player.x + (scale)*float32(math.Cos(player.direction))*1.5
	play.hat.Y3 = player.y + (scale)*float32(math.Sin(player.direction))*1.5
	play.wing1.X1 = player.body.X3
	play.wing1.Y1 = player.body.Y3
	play.wing1.X2 = player.wing1.X1 + scale*float32(0.8*math.Cos(player.direction))
	play.wing1.Y2 = player.wing1.Y1 + scale*float32(0.8*math.Sin(player.direction))
	play.wing1.X3 = player.wing1.X1 + scale*float32(0.8*math.Cos(player.direction-2*math.Pi/3))
	play.wing1.Y3 = player.wing1.Y1 + scale*float32(0.8*math.Sin(player.direction-2*math.Pi/3))
	play.wing2.X1 = player.body.X4
	play.wing2.Y1 = player.body.Y4
	play.wing2.X2 = player.wing2.X1 + scale*float32(0.8*math.Cos(player.direction))
	play.wing2.Y2 = player.wing2.Y1 + scale*float32(0.8*math.Sin(player.direction))
	play.wing2.X3 = player.wing2.X1 + scale*float32(0.8*math.Cos(player.direction+2*math.Pi/3))
	play.wing2.Y3 = player.wing2.Y1 + scale*float32(0.8*math.Sin(player.direction+2*math.Pi/3))
	play.body.draw()
	play.hat.draw()
	play.wing1.draw()
	play.wing2.draw()
}

func (play *Player) fire() {
	ball := Circle{player.hat.X3, player.hat.Y3, 0.04}
	var distfromplayer float64
	outer:
	for distfromplayer < 0.5 {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		distfromplayer = math.Sqrt(float64((ball.x-player.x)*(ball.x-player.x) + (ball.y-player.y)*(ball.y-player.y)))
		ball.x += 0.05 * float32(math.Cos(play.direction))
		ball.y += 0.05 * float32(math.Sin(play.direction))
		time.Sleep(time.Second / fps)
		ball.draw()
		for i := 0; i < 5; i++ {
			if float64(ball.r + troops[i].body.r) > math.Sqrt(float64((ball.x-troops[i].x)*(ball.x-troops[i].x)) + float64((ball.y-troops[i].y)*(ball.y-troops[i].y))) {
				troops[i].spawn()
				break outer
			}
		}
		draw()
	}
}
func (a *Triangle) getVao(colorpoints []float32) uint32 {
	return makeVao(a.getArray(), colorpoints)
}
func (a *Rect) getVao(colorpoints []float32) uint32 {
	return makeVao(a.getArray(), colorpoints)
}
func (a *Triangle) getArray() []float32 {
	return []float32{
		a.X1, a.Y1, 0, 0, 1, 0,
		a.X2, a.Y2, 0, 0, 1, 0,
		a.X3, a.Y3, 0, 0.5, 0.5, 0.5,
	}
}
func (c *Rect) draw() {
	gl.BindVertexArray(c.getVao(c.getColors()))
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.getArray())/3))
}

func (c *Rect) getColors() []float32 {
	return []float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
		1, 1, 1,
		1, 1, 0,
		0, 1, 1,
	}
}

func (c *Triangle) getColors() []float32 {
	return []float32{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

func (c *Triangle) draw() {
	gl.BindVertexArray(c.getVao(c.getColors()))
	//	gl.BindVertexArray(c.getVaoColor(1,0.5,1,1,1,1,1,1,1,1,1,1))
	//	gl.Color4ub(255, 255, 255, 255)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.getArray())/3))
}

type Rect struct {
	X1 float32
	Y1 float32
	X2 float32
	Y2 float32
	X3 float32
	Y3 float32
	X4 float32
	Y4 float32
}

func (a *Rect) getArray() []float32 {
	return []float32{
		a.X1, a.Y1, 0, 0, 1, 0,
		a.X2, a.Y2, 0, 0, 1, 0,
		a.X3, a.Y3, 0, 0, 1, 0,
		a.X3, a.Y3, 0, 0, 1, 0,
		a.X2, a.Y2, 0, 0, 1, 0,
		a.X4, a.Y4, 0, 0, 1, 0}
}
func main() {
	runtime.LockOSThread()

	window = initGlfw()
	window.SetCursorPosCallback(updatecursor)
	window.SetMouseButtonCallback(mouseButtonHandler)
	window.SetKeyCallback(keyHandler)
	defer glfw.Terminate()
	program = initOpenGL()
	for i := 0; i < len(troops); i++ {
		troops[i].spawn()
	}
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		time.Sleep(time.Second / fps)
		draw()
	}
}

func draw() {
	gl.UseProgram(program)
	player.draw(0.1)
	for i := 0; i < len(troops); i++ {
		troops[i].draw()
	}
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
	window.SetRefreshCallback(refresh)
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

func makeVao(points []float32, colorspoints []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STREAM_DRAW)
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 24, gl.PtrOffset(12))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 24, nil)
	gl.EnableVertexAttribArray(1)
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
func updatecursor(window *glfw.Window, x float64, y float64) {
	winheight, winwidth := window.GetFramebufferSize()
	// fmt.Println(winheight, winwidth)
	winglx, wingly := mgl64.ScreenToGLCoords(int(x), int(y), winheight, winwidth)
	player.direction = math.Atan2(wingly-float64(player.y), winglx-float64(player.x))
	// fmt.Println(map1(y, 0,float64(winheight), -1,1)-float64(player.x), (map1(x, 0,float64(winwidth), -1,1)-float64(player.y)))
}
func keyHandler(win *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyUp {
		player.x += 0.01 * float32(math.Cos(player.direction))
		player.y += 0.01 * float32(math.Sin(player.direction))
	}
	if key == glfw.KeyDown {
		player.x -= 0.01 * float32(math.Cos(player.direction))
		player.y -= 0.01 * float32(math.Sin(player.direction))
	}
	if key == glfw.KeyEscape {
		win.Destroy()
		os.Exit(0)
	}
}
func refresh(w *glfw.Window) {
	widthw, heightw := w.GetFramebufferSize()
	gl.Viewport(0, 0, int32(widthw), int32(heightw))
}

func map1(value, istart, istop, ostart, ostop float32) float32 {
	return ostart + (ostop-ostart)*((value-istart)/(istop-istart))
}
func mouseButtonHandler(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		player.fire()
	}
}

func readfile(filename string) string {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(s) + "\x00"
}
