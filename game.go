package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
)

type Game struct {
	tex          *TextureHandler
	gameLevels   *gameLevels
	camera       *raycaster.Camera
	scene        *ebiten.Image
	player       *Player
	vsync        bool
	fsr          float64
	opengl       bool
	headshot     *ebiten.Image
	screenWidth  int
	screenHeight int
	renderScale  float64
	width        int
	height       int
	minLightRGB  *color.NRGBA
	maxLightRGB  *color.NRGBA
}

func NewGame() *Game {
	fmt.Println("Creating game")
	g := new(Game)
	ebiten.SetWindowTitle("Game file")
	g.fsr = 4
	g.screenHeight = 600
	g.screenWidth = 800
	g.renderScale = 0.5
	if g.opengl {
		os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	}
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setVsyncEnabled(g.vsync)
	g.gameLevels = loadGameLevels()
	g.tex = NewTextureHandler(g.gameLevels)
	angleDegrees := 60.0
	g.player = NewPlayer(1.5, 1.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = 0.2
	g.player.CollisionHeight = 0.5
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.gameLevels.levelMaps[g.gameLevels.currentLevel], g.tex)
	g.camera.SetFloorTexture(getTextureFromFile("sky.png"))
	g.camera.SetSkyTexture(getTextureFromFile("sky.png"))
	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.camera.SetFovAngle(68, 1.0)
	g.setLightFalloff(-300)
	g.setGlobalIllumination(500)
	minLightRGB := &color.NRGBA{R: 15, G: 15, B: 15, A: 255}
	maxLightRGB := &color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	g.setLightRGB(minLightRGB, maxLightRGB)
	img, _, _ := ebitenutil.NewImageFromFile("./resources/headshot.png")
	g.headshot = img

	return (g)
}

func (g *Game) Run() {
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := g.screenWidth, g.screenHeight
	return w, h
}
func (g *Game) Update() error {
	g.handleInput()
	g.updateSprites()

	// handle player camera movement
	g.updatePlayerCamera(false)
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	sprites := g.gameLevels.levelMaps[g.gameLevels.currentLevel].sprites
	numSprites := len(sprites)
	raycastSprites := make([]raycaster.Sprite, numSprites)
	index := 0
	for sprite := range sprites {
		raycastSprites[index] = sprite
		index += 1
	}
	g.camera.Update(raycastSprites)
	g.camera.Draw(g.scene)
	op := &ebiten.DrawImageOptions{}
	if g.renderScale != 1.0 {
		op.Filter = ebiten.FilterNearest
		op.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
	}
	screen.DrawImage(g.scene, op)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(.15, .15)
	op.GeoM.Translate(0, 600-(float64(g.headshot.Bounds().Dy())*.15))
	screen.DrawImage(g.headshot, op)
	op2 := &ebiten.DrawImageOptions{}
	img, _, _ := ebitenutil.NewImageFromFile("./resources/revolver.png")
	op2.GeoM.Scale(2, 2)
	op2.GeoM.Translate(800-(float64(img.Bounds().Dx())*2), 600-(float64(img.Bounds().Dy())*2))

	screen.DrawImage(img, op2)

}

func (g *Game) setResolution(screenWidth, screenHeight int) {
	g.screenWidth, g.screenHeight = screenWidth, screenHeight
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g.setRenderScale(g.renderScale)
}

func (g *Game) setRenderScale(renderScale float64) {
	g.renderScale = renderScale
	g.width = int(math.Floor(float64(g.screenWidth) * g.renderScale))
	g.height = int(math.Floor(float64(g.screenHeight) * g.renderScale))
	if g.camera != nil {
		g.camera.SetViewSize(g.width, g.height)
	}
	g.scene = ebiten.NewImage(g.width, g.height)
}

func (g *Game) updatePlayerCamera(forceUpdate bool) {
	if !g.player.Moved && !forceUpdate {
		// only update camera position if player moved or forceUpdate set
		return
	}

	// reset player moved flag to only update camera when necessary
	g.player.Moved = false

	g.camera.SetPosition(g.player.Position.Copy())
	g.camera.SetPositionZ(g.player.CameraZ)
	g.camera.SetHeadingAngle(g.player.Angle)
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.camera.SetFovAngle(fovDegrees, 1.0)
}
func (g *Game) setLightFalloff(lightFalloff float64) {
	g.camera.SetLightFalloff(lightFalloff)
}
func (g *Game) setGlobalIllumination(globalIllumination float64) {
	g.camera.SetGlobalIllumination(globalIllumination)
}

func (g *Game) setLightRGB(minLightRGB, maxLightRGB *color.NRGBA) {
	g.minLightRGB = minLightRGB
	g.maxLightRGB = maxLightRGB
	g.camera.SetLightRGB(*g.minLightRGB, *g.maxLightRGB)
}
func (g *Game) Rotate(rSpeed float64) {
	g.player.Angle += rSpeed

	for g.player.Angle > geom.Pi {
		g.player.Angle = g.player.Angle - geom.Pi2
	}
	for g.player.Angle <= -geom.Pi {
		g.player.Angle = g.player.Angle + geom.Pi2
	}

	g.player.Moved = true
}
func (g *Game) Move(mSpeed float64) {
	moveLine := geom.LineFromAngle(g.player.Position.X, g.player.Position.Y, g.player.Angle, mSpeed)

	newPos, _, _ := g.getValidMove(g.player.Entity, moveLine.X2, moveLine.Y2, g.player.PositionZ, true)
	if !newPos.Equals(g.player.Pos()) {
		g.player.Position = newPos
		g.player.Moved = true
	}
}
func (g *Game) setVsyncEnabled(enableVsync bool) {
	g.vsync = enableVsync
	ebiten.SetVsyncEnabled(enableVsync)
}
func (g *Game) updateSprites() {
	// Testing animated sprite movement
	sprites := g.gameLevels.levelMaps[g.gameLevels.currentLevel].sprites
	for s := range sprites {
		if s.Velocity != 0 {
			vLine := geom.LineFromAngle(s.Position.X, s.Position.Y, s.Angle, s.Velocity)

			xCheck := vLine.X2
			yCheck := vLine.Y2
			zCheck := s.PositionZ

			newPos, isCollision, _ := g.getValidMove(s.Entity, xCheck, yCheck, zCheck, false)
			if isCollision {
				// for testing purposes, letting the sample sprite ping pong off walls in somewhat random direction
				s.Angle = randFloat(-math.Pi, math.Pi)
				s.Velocity = randFloat(0.01, 0.03)
			} else {
				s.Position = newPos
			}
		}
		s.Update(g.player.Position)
	}
}
func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
