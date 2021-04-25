package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

var (
	player *Shape
)

func InitSprites() {
	player = NewShape(mgl32.Ident4(), program, 
	P(0.5,1,1),
	P()
)
}
