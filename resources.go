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

// loadContent will be called once per game and is the place to load
// all of your content.
func (g *Game) loadContent() {

	// TODO: make resource management better

	// load wall textures
	g.tex.textures[0] = getTextureFromFile("stone.png")
	g.tex.textures[1] = getTextureFromFile("left_bot_house.png")
	g.tex.textures[2] = getTextureFromFile("right_bot_house.png")
	g.tex.textures[3] = getTextureFromFile("left_top_house.png")
	g.tex.textures[4] = getTextureFromFile("right_top_house.png")

	// separating sprites out a bit from wall textures
	g.tex.textures[5] = getSpriteFromFile("large_rock.png")
	g.tex.textures[6] = getSpriteFromFile("tree_09.png")
	g.tex.textures[7] = getSpriteFromFile("tree_10.png")
	g.tex.textures[8] = getSpriteFromFile("tree_14.png")
	g.tex.floorTex = getRGBAFromFile("grass.png")
	g.tex.ceilingTex = getRGBAFromFile("ceiling.png")
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

func (g *Game) loadSprites() {
	g.sprites = make(map[*Sprite]struct{}, 128)

	// colors for minimap representation
	brown := color.RGBA{47, 40, 30, 196}
	green := color.RGBA{27, 37, 7, 196}
	orange := color.RGBA{69, 30, 5, 196}

	// rock that can be jumped over but not walked through
	rockImg := g.tex.textures[8]
	rockWidth, rockHeight := rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
	rockScale := 0.4
	rockPxRadius, rockPxHeight := 24.0, 35.0
	rockCollisionRadius := (rockScale * rockPxRadius) / float64(rockWidth)
	rockCollisionHeight := (rockScale * rockPxHeight) / float64(rockHeight)
	rock := NewSprite(8.0, 5.5, rockScale, rockImg, brown, raycaster.AnchorBottom, rockCollisionRadius, rockCollisionHeight)
	g.addSprite(rock)

	// testing sprite scaling
	testScale := 0.5
	g.addSprite(NewSprite(10.5, 2.5, testScale, g.tex.textures[5], green, raycaster.AnchorBottom, 0, 0))

	// // line of trees for testing in front of initial view
	// Setting CollisionRadius=0 to disable collision against small trees
	g.addSprite(NewSprite(19.5, 11.5, 1.0, g.tex.textures[6], brown, raycaster.AnchorBottom, 0, 0))
	g.addSprite(NewSprite(17.5, 11.5, 1.0, g.tex.textures[7], orange, raycaster.AnchorBottom, 0, 0))
	g.addSprite(NewSprite(15.5, 11.5, 1.0, g.tex.textures[8], green, raycaster.AnchorBottom, 0, 0))

}

func (g *Game) addSprite(sprite *Sprite) {
	g.sprites[sprite] = struct{}{}
}
