package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go"
	"github.com/harbdog/raycaster-go/geom"
	"github.com/spf13/viper"
)

const (
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	texWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	clipDistance = 0.1
)

type Game struct {
	tex                  *TextureHandler
	screenWidth          int
	screenHeight         int
	camera               *raycaster.Camera
	scene                *ebiten.Image
	mapObj               *Map
	renderScale          float64
	height               int
	width                int
	mapWidth             int
	mapHeight            int
	sprites              map[*Sprite]struct{}
	fovDegrees           float64
	fovDepth             float64
	zoomFovDepth         float64
	renderDistance       float64
	lightFalloff         float64
	globalIllumination   float64
	minLightRGB          *color.NRGBA
	maxLightRGB          *color.NRGBA
	player               *Player
	collisionMap         []geom.Line
	vsync                bool
	fsr                  float64
	opengl               bool
	showSpriteBoxes      bool
	initRenderFloorTex   bool
	initRenderCeilingTex bool
}

func NewGame() *Game {
	fmt.Println("Creating game")
	g := new(Game)
	g.initConfig()
	ebiten.SetWindowTitle("Game file")
	if g.opengl {
		os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	}
	g.setResolution(g.screenWidth, g.screenHeight)
	g.setRenderScale(g.renderScale)
	g.setVsyncEnabled(g.vsync)
	g.mapObj = NewMap()
	g.tex = NewTextureHandler(g.mapObj, 32)
	worldMap := g.mapObj.worldMap
	g.mapWidth = len(worldMap)
	g.mapHeight = len(worldMap[0])
	g.loadContent()
	angleDegrees := 60.0
	g.player = NewPlayer(8.5, 3.5, geom.Radians(angleDegrees), 0)
	g.player.CollisionRadius = clipDistance
	g.player.CollisionHeight = 0.5
	g.loadSprites()
	g.camera = raycaster.NewCamera(g.width, g.height, texWidth, g.mapObj, g.tex)
	g.camera.SetFloorTexture(getTextureFromFile("floor.png"))
	g.camera.SetSkyTexture(getTextureFromFile("sky.png"))
	// initialize camera to player position
	g.updatePlayerCamera(true)
	g.setFovAngle(g.fovDegrees)
	g.fovDepth = g.camera.FovDepth()

	g.zoomFovDepth = 2.0
	g.setLightFalloff(-300)
	g.setGlobalIllumination(500)
	minLightRGB := &color.NRGBA{R: 15, G: 15, B: 15, A: 255}
	maxLightRGB := &color.NRGBA{R: 180, G: 180, B: 180, A: 255}
	g.setLightRGB(minLightRGB, maxLightRGB)

	return (g)
}
func (g *Game) initConfig() {
	viper.SetConfigName("demo-config")
	viper.SetConfigType("json")

	// setup environment variable with DEMO as prefix (e.g. "export DEMO_SCREEN_VSYNC=false")
	viper.SetEnvPrefix("demo")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	userHomePath, _ := os.UserHomeDir()
	if userHomePath != "" {
		userHomePath = userHomePath + "/.raycaster-go-demo"
		viper.AddConfigPath(userHomePath)
	}
	viper.AddConfigPath(".")

	// set default config values
	viper.SetDefault("debug", false)
	viper.SetDefault("showSpriteBoxes", false)
	viper.SetDefault("screen.fullscreen", false)
	viper.SetDefault("screen.vsync", true)
	viper.SetDefault("screen.fsr", 4.0)
	viper.SetDefault("screen.renderDistance", -1)
	viper.SetDefault("screen.renderFloor", true)
	viper.SetDefault("screen.renderCeiling", true)
	viper.SetDefault("screen.fovDegrees", 68)

	viper.SetDefault("screen.width", 1024)
	viper.SetDefault("screen.height", 768)
	viper.SetDefault("screen.renderScale", 1.0)

	if runtime.GOOS == "windows" {
		// default windows to opengl for better performance
		viper.SetDefault("screen.opengl", true)
	}

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Print(err)
	}

	// get config values
	g.screenWidth = viper.GetInt("screen.width")
	g.screenHeight = viper.GetInt("screen.height")
	g.fovDegrees = viper.GetFloat64("screen.fovDegrees")
	g.renderScale = viper.GetFloat64("screen.renderScale")
	g.vsync = viper.GetBool("screen.vsync")
	g.fsr = viper.GetFloat64("screen.fsr")
	g.opengl = viper.GetBool("screen.opengl")
	g.renderDistance = viper.GetFloat64("screen.renderDistance")
	g.initRenderFloorTex = viper.GetBool("screen.renderFloor")
	g.initRenderCeilingTex = viper.GetBool("screen.renderCeiling")
	g.showSpriteBoxes = viper.GetBool("showSpriteBoxes")
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
	numSprites := len(g.sprites)
	raycastSprites := make([]raycaster.Sprite, numSprites)
	index := 0
	for sprite := range g.sprites {
		raycastSprites[index] = sprite
		index += 1
	}
	for depth := 20; depth >= 0; depth-- {
		g.camera.Update(raycastSprites, depth)
		g.camera.Draw(g.scene, depth)
		op := &ebiten.DrawImageOptions{}
		if g.renderScale != 1.0 {
			op.Filter = ebiten.FilterNearest
			op.GeoM.Scale(1/g.renderScale, 1/g.renderScale)
		}
		screen.DrawImage(g.scene, op)
	}

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
	g.camera.SetPitchAngle(g.player.Pitch)
}

func (g *Game) setFovAngle(fovDegrees float64) {
	g.fovDegrees = fovDegrees
	g.camera.SetFovAngle(fovDegrees, 1.0)
}
func (g *Game) setLightFalloff(lightFalloff float64) {
	g.lightFalloff = lightFalloff
	g.camera.SetLightFalloff(g.lightFalloff)
}
func (g *Game) setGlobalIllumination(globalIllumination float64) {
	g.globalIllumination = globalIllumination
	g.camera.SetGlobalIllumination(g.globalIllumination)
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
	for s := range g.sprites {
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
