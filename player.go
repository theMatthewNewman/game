package main

import (
	"image/color"

	"github.com/harbdog/raycaster-go/geom"
)

type Player struct {
	*Entity
	CameraZ float64
	Moved   bool
}

func NewPlayer(x, y, angle, pitch float64) *Player {
	p := &Player{
		Entity: &Entity{
			Position:  &geom.Vector2{X: x, Y: y},
			PositionZ: 0,
			Angle:     angle,
			Pitch:     pitch,
			Velocity:  0,
			MapColor:  color.RGBA{255, 0, 0, 255},
		},
		CameraZ: 0.5,
		Moved:   false,
	}

	return p
}
