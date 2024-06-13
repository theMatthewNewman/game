package main

import (
	"embed"
	"image"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/harbdog/raycaster-go"
)

//go:embed resources
var embedded embed.FS

const (
	texWidth = 256
)

// loadContent will be called once per game and is the place to load
// all of your content.
func (t *TextureHandler) loadTextureFiles() {

	t.wallTextures[0] = getTextureFromFile("fence.png")
	t.wallTextures[1] = getTextureFromFile("woodfloor.png")
	t.wallTextures[2] = getTextureFromFile("slab.png")
	t.wallTextures[3] = getTextureFromFile("wallpaper.png")
	t.wallTextures[4] = getTextureFromFile("window1.png")

	t.floorAndCeilingTextures[0] = getRGBAFromFile("stone.png")
	t.floorAndCeilingTextures[1] = getRGBAFromFile("woodfloor.png")
	t.floorAndCeilingTextures[2] = getRGBAFromFile("grass.png")
	t.floorAndCeilingTextures[3] = getRGBAFromFile("woodfloor.png")
	t.floorAndCeilingTextures[4] = getRGBAFromFile("woodfloor.png")
	t.floorAndCeilingTextures[5] = getRGBAFromFile("carpet1.png")
	t.floorAndCeilingTextures[6] = getRGBAFromFile("carpet2.png")
	t.floorAndCeilingTextures[7] = getRGBAFromFile("carpet3.png")
	t.floorAndCeilingTextures[8] = getRGBAFromFile("carpet4.png")

	t.spriteTextures[0] = getSpriteFromFile("large_rock.png")
	t.spriteTextures[1] = getSpriteFromFile("couch.png")
	t.spriteTextures[2] = getSpriteFromFile("couch.png")
	t.spriteTextures[3] = getSpriteFromFile("couch.png")
}

func newImageFromFile(path string) (*ebiten.Image, image.Image, error) {
	f, err := embedded.Open(filepath.ToSlash(path))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	eb, im, err := ebitenutil.NewImageFromReader(f)
	return eb, im, err
}

func newScaledImageFromFile(path string, scale float64) (*ebiten.Image, image.Image, error) {
	eb, im, err := newImageFromFile(path)
	if err != nil {
		return eb, im, err
	}

	if scale == 1.0 {
		return eb, im, err
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)

	scaledWidth, scaledHeight := float64(eb.Bounds().Dx())*scale, float64(eb.Bounds().Dy())*scale
	scaledImage := ebiten.NewImage(int(scaledWidth), int(scaledHeight))
	scaledImage.DrawImage(eb, op)

	return scaledImage, scaledImage, err
}

func getRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	_, tex, err := newImageFromFile("resources/" + texFile)
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, texWidth, texWidth))
		// convert into RGBA format
		for x := 0; x < texWidth; x++ {
			for y := 0; y < texWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func getTextureFromFile(texFile string) *ebiten.Image {
	eImg, _, err := newImageFromFile("resources/" + texFile)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func getSpriteFromFile(sFile string) *ebiten.Image {
	eImg, _, err := newImageFromFile("resources/" + sFile)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func (t *TextureHandler) loadSprites() {
	currentLevel := t.gameLevels.levelMaps[t.gameLevels.currentLevel]
	currentLevel.sprites = make(map[*Sprite]struct{}, 128)

	// colors for minimap representation
	brown := color.RGBA{47, 40, 30, 196}
	green := color.RGBA{27, 37, 7, 196}
	orange := color.RGBA{69, 30, 5, 196}

	// rock that can be jumped over but not walked through
	rockImg := t.spriteTextures[0]
	rockWidth, rockHeight := rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
	rockScale := 0.4
	rockPxRadius, rockPxHeight := 24.0, 35.0
	rockCollisionRadius := (rockScale * rockPxRadius) / float64(rockWidth)
	rockCollisionHeight := (rockScale * rockPxHeight) / float64(rockHeight)
	rock := NewSprite(8.0, 5.5, rockScale, rockImg, brown, raycaster.AnchorBottom, rockCollisionRadius, rockCollisionHeight)
	currentLevel.addSprite(rock)
	rockImg = t.spriteTextures[1]
	rockWidth, rockHeight = rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
	rockScale = 0.4
	rockPxRadius, rockPxHeight = 24.0, 35.0
	rockCollisionRadius = (rockScale * rockPxRadius) / float64(rockWidth)
	rockCollisionHeight = (rockScale * rockPxHeight) / float64(rockHeight)
	rock = NewSprite(8.5, 5.5, rockScale, rockImg, brown, raycaster.AnchorBottom, rockCollisionRadius, rockCollisionHeight)
	currentLevel.addSprite(rock)

	// // line of trees for testing in front of initial view
	// Setting CollisionRadius=0 to disable collision against small trees
	currentLevel.addSprite(NewSprite(19.5, 11.5, 1.0, t.spriteTextures[1], brown, raycaster.AnchorBottom, 0, 0))
	currentLevel.addSprite(NewSprite(17.5, 11.5, 1.0, t.spriteTextures[2], orange, raycaster.AnchorBottom, 0, 0))
	currentLevel.addSprite(NewSprite(15.5, 11.5, 1.0, t.spriteTextures[3], green, raycaster.AnchorBottom, 0, 0))

}

func (m *Map) addSprite(sprite *Sprite) {
	m.sprites[sprite] = struct{}{}
}
