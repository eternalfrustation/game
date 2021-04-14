package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

var (
	player *Shape
)

func InitSprites() {
	Player = NewShape(mgl32.Ident4(), program)
}
