package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

var (
	Player = new(Shape)
)

func InitSprites() {
	Player.Pts = append(Player.Pts, &Point{P: mgl32.Vec3{}})
}
