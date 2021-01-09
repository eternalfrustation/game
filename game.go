package main

import (
	"fmt"
	"unsafe"
	//"github.com/go-gl/gltext"
	"log"
	"math"

	"github.com/hajimehoshi/oto"

	"math/rand"
	"os"

	"runtime"
	// "strconv"
	"io"
	"strings"
	"sync"
	"time"

	// "path/filepath"
	"io/ioutil"

	"github.com/eternalfrustation/fontgl"

	// "unsafe"
	"flag"

	"github.com/go-gl/gl/v4.1-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl64"
)

const (
	width  = 500
	height = 500
	fps    = 60
)

var (
	program              uint32
	window               *glfw.Window
	player               = *new(Player)
	vertexShaderSource   = readfile("vert.glsl")
	fragmentShaderSource = readfile("frag.glsl")
	troops               = *new([5]Enemy)
	maxspeed             = float32(0.0125)
	sampleRate           = flag.Int("samplerate", 44100, "sample rate")
	channelNum           = flag.Int("channelnum", 2, "number of channel")
	bitDepthInBytes      = flag.Int("bitdepthinbytes", 2, "bit depth in bytes")
	c, _                 = oto.NewContext(*sampleRate, *channelNum, *bitDepthInBytes, 4096)
	wg                   sync.WaitGroup
	IsFired              bool
	firefile, _          = os.Open("Laser_Shoot.wav")
	hurtfile, _          = os.Open("Hurt.wav")
	explosionfile, _     = os.Open("Explosion.wav")
)

type SineWave struct {
	freq   float64
	length int64
	pos    int64

	remaining []byte
}

func NewSineWave(freq float64, duration time.Duration) *SineWave {
	l := int64(*channelNum) * int64(*bitDepthInBytes) * int64(*sampleRate) * int64(duration) / int64(time.Second)
	l = l / 4 * 4
	return &SineWave{
		freq:   freq,
		length: l,
	}
}

func (s *SineWave) Read(buf []byte) (int, error) {
	if len(s.remaining) > 0 {
		n := copy(buf, s.remaining)
		s.remaining = s.remaining[n:]
		return n, nil
	}

	if s.pos == s.length {
		return 0, io.EOF
	}

	eof := false
	if s.pos+int64(len(buf)) > s.length {
		buf = buf[:s.length-s.pos]
		eof = true
	}

	var origBuf []byte
	if len(buf)%4 > 0 {
		origBuf = buf
		buf = make([]byte, len(origBuf)+4-len(origBuf)%4)
	}

	length := float64(*sampleRate) / float64(s.freq)

	num := (*bitDepthInBytes) * (*channelNum)
	p := s.pos / int64(num)
	switch *bitDepthInBytes {
	case 1:
		for i := 0; i < len(buf)/num; i++ {
			const max = 127
			b := int(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < *channelNum; ch++ {
				buf[num*i+ch] = byte(b + 128)
			}
			p++
		}
	case 2:
		for i := 0; i < len(buf)/num; i++ {
			const max = 32767
			b := int16(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < *channelNum; ch++ {
				buf[num*i+2*ch] = byte(b)
				buf[num*i+1+2*ch] = byte(b >> 8)
			}
			p++
		}
	}

	s.pos += int64(len(buf))

	n := len(buf)
	if origBuf != nil {
		n = copy(origBuf, buf)
		s.remaining = buf[n:]
	}

	if eof {
		return n, io.EOF
	}
	return n, nil
}

func playsound(context *oto.Context, freq float64, duration time.Duration) error {
	p := context.NewPlayer()
	s := NewSineWave(freq, duration)
	if _, err := io.Copy(p, s); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}
	return nil
}

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
	//fmt.Println(circleLineColl(player.hat.X3, player.hat.Y3, player.hat.X1, player.hat.Y1, e.x, e.y, e.size) || circleLineColl(player.hat.X3, player.hat.Y3, player.hat.X2, player.hat.Y2, e.x, e.y, e.size) || circleLineColl(player.body.X1, player.body.Y1, player.body.X3, player.body.Y3, e.x, e.y, e.size) || circleLineColl(player.body.X2, player.body.Y2, player.body.X4, player.body.Y3, e.x, e.y, e.size))
	if circleLineColl(player.hat.X3, player.hat.Y3, player.hat.X1, player.hat.Y1, e.x, e.y, e.size) || circleLineColl(player.hat.X3, player.hat.Y3, player.hat.X2, player.hat.Y2, e.x, e.y, e.size) || circleLineColl(player.body.X1, player.body.Y1, player.body.X3, player.body.Y3, e.x, e.y, e.size) || circleLineColl(player.body.X2, player.body.Y2, player.body.X4, player.body.Y3, e.x, e.y, e.size) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			playersound := c.NewPlayer()
			file, err := os.Open("exp.wav")
			if err != nil {
				panic(err)
			}
			fmt.Println(file)
			fmt.Println(io.Copy(playersound, file))
			os.Exit(0)
		}()
		// TODO
	}
}

func circleLineColl(x1, y1, x2, y2, cx, cy, r float32) bool {
	m := (y1 - y2) / (x1 - x2)
	closestx := (x1 + m*cx + m*m*x1 + m*cy) / (m*m + 1)
	closesty := (y1 + m*cy + m*m*y1 + m*cx) / (m*m + 1)
	return (distancesq(cx, cy, closestx, closesty) < r*r) && (distancesq(x1, y1, closestx, closesty)+distancesq(x2, y2, closestx, closesty) <= distancesq(x1, y1, x2, y2))
}

func distancesq(x1, y1, x2, y2 float32) float32 {
	return (x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)
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
	for ii := 0; ii < num_segments; ii++ {
		circleslice = append(circleslice, cx, cy, 0, 0, 0, 0, 0, 0, 1)
		circleslice = append(circleslice, x+cx, y+cy, 0, 0.8, 0.8, 0.8, 0, 0, 1)

		//apply the rotation matrix
		t = x
		x = float32(c)*x - float32(s)*y
		y = float32(s)*t + float32(c)*y
	}
	gl.BindVertexArray(makeVao(circleslice))
	gl.DrawArrays(gl.LINES, 0, int32(len(circleslice))/3)

}

type Player struct {
	body      Rect
	hat       Triangle
	wing1     Triangle
	wing2     Triangle
	x         float32
	y         float32
	v         float32
	direction float64
	hp        int
}

func (play *Player) draw(scale float32) {
	if player.hp < 1 {
		os.Exit(0)
	}
	play.x += play.v * float32(math.Cos(player.direction))
	play.y += play.v * float32(math.Sin(player.direction))
	if play.v > 0 {
		play.v -= 0.001
	}
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
	balldir := play.direction
	wg.Add(1)
	go func() {
		defer wg.Done()
		playersound := c.NewPlayer()
		file, err := os.Open("ls.wav")
		if err != nil {
			panic(err)
		}
		fmt.Println(file)
		fmt.Println(io.Copy(playersound, file))
	}()
outer:
	for distfromplayer < 0.5 {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		distfromplayer = math.Sqrt(float64((ball.x-player.x)*(ball.x-player.x) + (ball.y-player.y)*(ball.y-player.y)))
		ball.x += 0.05 * float32(math.Cos(balldir))
		ball.y += 0.05 * float32(math.Sin(balldir))
		time.Sleep(time.Second / fps)
		ball.draw()
		for i := 0; i < 5; i++ {
			if float64(ball.r+troops[i].body.r) > math.Sqrt(float64((ball.x-troops[i].x)*(ball.x-troops[i].x))+float64((ball.y-troops[i].y)*(ball.y-troops[i].y))) {
				troops[i].spawn()
				draw()
				break outer
			}
		}
		draw()
	}
}
func (a *Triangle) getVao() uint32 {
	return makeVao(a.getArray())
}
func (a *Rect) getVao() uint32 {
	return makeVao(a.getArray())
}
func (a *Triangle) getArray() []float32 {
	return []float32{
		a.X1, a.Y1, 0, 0, 1, 0, 0, 0, 1,
		a.X2, a.Y2, 0, 0, 1, 0, 0, 0, 1,
		a.X3, a.Y3, 0, 0.5, 0.5, 0.5, 0, 0, 1,
	}
}
func (c *Rect) draw() {
	gl.BindVertexArray(c.getVao())
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(c.getArray())/3))
}

func (c *Triangle) draw() {
	gl.BindVertexArray(c.getVao())
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
		a.X1, a.Y1, 0, 0, 1, 0, 0, 0, 1,
		a.X2, a.Y2, 0, 0, 1, 0, 0, 0, 1,
		a.X3, a.Y3, 0, 0, 1, 0, 0, 0, 1,
		a.X3, a.Y3, 0, 0, 1, 0, 0, 0, 1,
		a.X2, a.Y2, 0, 0, 1, 0, 0, 0, 1,
		a.X4, a.Y4, 0, 0, 1, 0, 0, 0, 1}
}
func main() {
	runtime.LockOSThread()
	window = initGlfw()
	window.SetCursorPosCallback(updatecursor)
	window.SetMouseButtonCallback(mouseButtonHandler)
	window.SetKeyCallback(keyHandler)
	defer glfw.Terminate()
	program = initOpenGL()
	player.hp = 100
	for i := 0; i < len(troops); i++ {
		troops[i].spawn()
	}
	//	font.LoadGlyph(nil, sfnt.)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	// code from here
	gl.LineWidth(1)
	gl.PointSize(50)
	fnt := fontgl.Setup("font/font.ttf")
	Lines, _ := fontgl.GetTriText("H", fnt, -1, -1)
	//	fmt.Println(Lines)
	fmt.Println("after")
	gl.DebugMessageCallback(glprinterr, nil)
	Lvao := makeVao(Lines)
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		time.Sleep(time.Second / fps)

		gl.UseProgram(program)
		//	fmt.Println("Lines")
		gl.BindVertexArray(Lvao)
		gl.DrawArrays(gl.LINES, 0, int32(len(Lines)/3))
		//	fmt.Println("Points")
		draw()
		//	drawstring("hello", 0, 0)

		glfw.PollEvents()
		window.SwapBuffers()
	}
}
func glprinterr(source, gltype , id, severity uint32, length int32, message string, userparam unsafe.Pointer) {
	fmt.Println(source, gltype, id, severity, length, message)
}
func draw() {
	player.draw(0.1)
	for i := 0; i < len(troops); i++ {
		troops[i].draw()
	}
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

func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(&points[0]), gl.DYNAMIC_DRAW)
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 36, gl.PtrOffset(12))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 36, nil)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 36, gl.PtrOffset(24))
	gl.EnableVertexAttribArray(2)
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
	if key == glfw.KeyUp && player.v < maxspeed {
		player.v += 0.01
	}
	if key == glfw.KeyDown && player.v > -maxspeed {
		player.v -= 0.01
	}
	if key == glfw.KeyEscape {
		win.Destroy()
		os.Exit(0)
	}
}
func refresh(w *glfw.Window) {
	widthw, heightw := w.GetFramebufferSize()
	gl.Viewport(0, 0, int32(widthw), int32(heightw))
	gl.MatrixMode(gl.PROJECTION_MATRIX)
	gl.LoadIdentity()
	newmat := mgl64.Perspective(mgl64.DegToRad(180), float64(widthw)/float64(heightw), 0.1, 100.0)
	gl.MultMatrixd(&newmat[0])
}

func map1(value, istart, istop, ostart, ostop float32) float32 {
	return ostart + (ostop-ostart)*((value-istart)/(istop-istart))
}
func mouseButtonHandler(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press && !IsFired {
		IsFired = true
		player.fire()
		IsFired = false
	}
}

func readfile(filename string) string {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(s) + "\x00"
}
