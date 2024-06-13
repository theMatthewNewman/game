package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	numFloorAndCeilingTextures = 9
	numWallTextures            = 5
	numSpriteTextures          = 5
)

type TextureHandler struct {
	gameLevels              *gameLevels
	wallTextures            []*ebiten.Image
	spriteTextures          []*ebiten.Image
	floorAndCeilingTextures []*image.RGBA
}

func NewTextureHandler(gameLevels *gameLevels) *TextureHandler {
	t := &TextureHandler{
		gameLevels:              gameLevels,
		wallTextures:            make([]*ebiten.Image, numWallTextures),
		floorAndCeilingTextures: make([]*image.RGBA, numFloorAndCeilingTextures),
		spriteTextures:          make([]*ebiten.Image, numSpriteTextures),
	}
	t.loadTextureFiles()
	t.loadSprites()
	return t
}

func (t *TextureHandler) TextureAt(x int, y int, levelNum int, side int) *ebiten.Image {
	texNum := -1

	mapLevel := t.gameLevels.levelMaps[t.gameLevels.currentLevel]

	if x >= 0 && x < mapLevel.xLength && y >= 0 && y < mapLevel.yLength {
		texNum = mapLevel.wallMaps[levelNum][x][y] - 1 // 1 subtracted from it so that texture 0 can be used
	}
	if texNum < 0 {
		return nil
	}
	return t.wallTextures[texNum]
}

func (t *TextureHandler) FloorTextureAt(x, y, z int) *image.RGBA {
	return t.floorAndCeilingTextures[0]
	texNum := -1

	mapLevel := t.gameLevels.levelMaps[t.gameLevels.currentLevel]

	if x >= 0 && x < mapLevel.xLength && y >= 0 && y < mapLevel.yLength {
		texNum = mapLevel.floorMap[x][y] - 1 // 1 subtracted from it so that texture 0 can be used
	}
	if texNum < 0 {
		return nil
	}
	return t.floorAndCeilingTextures[texNum]
}

func (t *TextureHandler) CeilingTextureAt(x, y, z int) *image.RGBA {
	texNum := -1

	mapLevel := t.gameLevels.levelMaps[t.gameLevels.currentLevel]

	if x >= 0 && x < mapLevel.xLength && y >= 0 && y < mapLevel.yLength {
		texNum = mapLevel.ceilingMap[x][y] - 1 // 1 subtracted from it so that texture 0 can be used
	}
	if texNum < 0 {
		return nil
	}
	return t.floorAndCeilingTextures[texNum]
}

func (t *TextureHandler) MidTextureAt(x, y, z int) *image.RGBA {
	if x == 5 && y == 5 {
		return t.floorAndCeilingTextures[0]
	}
	return nil
}
