//go:generate statik -src=./assets -include=*.png

package main

import (
	"errors"
	"image"
	"image/png"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	_ "github.com/jatekalkotok/lunar-defence/statik"
	"github.com/rakyll/statik/fs"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Lunar Defence")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)

	gameWidth, gameHeight := 1280, 960
	rand.Seed(time.Now().UnixNano())

	moonImage := loadImage("/moon.png")
	moon := &Moon{
		Image:  moonImage,
		Op:     &ebiten.DrawImageOptions{},
		Radius: float64(moonImage.Bounds().Dx()) / 2,
	}

	earthImage := loadImage("/earth.png")
	earth := &Earth{
		Image:  earthImage,
		Op:     &ebiten.DrawImageOptions{},
		Radius: float64(earthImage.Bounds().Dx()) / 2,
		Center: image.Point{gameWidth / 2, gameHeight / 2},
	}

	asteroidImage := loadImage("/asteroid.png")
	asteroid := &Asteroid{
		Image:    asteroidImage,
		Op:       &ebiten.DrawImageOptions{},
		Radius:   float64(asteroidImage.Bounds().Dx()) / 2,
		Angle:    rand.Float64() * math.Pi * 2,
		Distance: earth.Radius * 2,
	}

	crosshairImage := loadImage("/crosshair.png")
	crosshair := &Crosshair{
		Image:  crosshairImage,
		Op:     &ebiten.DrawImageOptions{},
		Radius: float64(crosshairImage.Bounds().Dx()) / 2,
	}

	game := &Game{
		Width:     gameWidth,
		Height:    gameHeight,
		Rotation:  0,
		Moon:      moon,
		Earth:     earth,
		Asteroid:  asteroid,
		Crosshair: crosshair,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Width     int
	Height    int
	Rotation  float64
	Moon      *Moon
	Earth     *Earth
	Asteroid  *Asteroid
	Crosshair *Crosshair
}

// Update calculates game logic
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return errors.New("game quit by player")
	}

	if g.Asteroid.Distance <= 0 {
		return nil
	}

	g.Rotation = g.Rotation - 0.02
	g.Asteroid.Distance = g.Asteroid.Distance - 1

	// Update object positions
	g.Earth.Update(g)
	g.Moon.Update(g)
	g.Asteroid.Update(g)
	g.Crosshair.Update()

	return nil
}

// Draw handles rendering the sprites
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.Earth.Image, g.Earth.Op)
	screen.DrawImage(g.Moon.Image, g.Moon.Op)
	screen.DrawImage(g.Asteroid.Image, g.Asteroid.Op)
	screen.DrawImage(g.Crosshair.Image, g.Crosshair.Op)
	// debug(screen, g)
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

// Moon is moon
type Moon struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Radius float64
}

// Update recalculates moon position
func (o Moon) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-g.Earth.Radius-o.Radius*2,
		-g.Earth.Radius-o.Radius*2,
	)
	o.Op.GeoM.Rotate(g.Rotation / 3)
	o.Op.GeoM.Translate(g.Earth.Pt())
}

// Earth is earth
type Earth struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Radius float64
	Center image.Point
}

// Update repositions Earth
func (o Earth) Update(g *Game) {
	o.Op.GeoM.Reset()
	o.Op.GeoM.Translate(
		-o.Radius,
		-o.Radius,
	)
	o.Op.GeoM.Rotate(g.Rotation)
	o.Op.GeoM.Translate(o.Pt())
}

// Pt is a shortcut for the Earth's X and Y coordinates
func (o Earth) Pt() (X, Y float64) {
	return float64(o.Center.X), float64(o.Center.Y)
}

// Asteroid is asteroid
type Asteroid struct {
	Image    *ebiten.Image
	Op       *ebiten.DrawImageOptions
	Radius   float64
	Angle    float64
	Distance float64
}

// Update recalculates Asteroid position
func (o Asteroid) Update(g *Game) {
	const RotationSpeed float64 = 3
	o.Op.GeoM.Reset()

	// Spin the asteroid
	o.Op.GeoM.Translate(-o.Radius, -o.Radius)
	o.Op.GeoM.Rotate(g.Rotation * RotationSpeed)

	// Move it back to where it was because maths is hard
	o.Op.GeoM.Translate(o.Radius, o.Radius)

	// Positions it at correct distance for angle correction
	o.Op.GeoM.Translate(
		-g.Earth.Radius+o.Radius*2-o.Distance,
		-g.Earth.Radius+o.Radius*2-o.Distance,
	)

	// Turn to correct angle
	o.Op.GeoM.Rotate(o.Angle)

	// Move post-rotation centre to match Earth's centre
	o.Op.GeoM.Translate(g.Earth.Pt())
}

// The Crosshair is a target showing where the the player will shoot
type Crosshair struct {
	Image  *ebiten.Image
	Op     *ebiten.DrawImageOptions
	Radius float64
}

// Update recalculates the crosshair position
func (o Crosshair) Update() {
	o.Op.GeoM.Reset()
	mx, my := ebiten.CursorPosition()
	o.Op.GeoM.Translate(
		float64(mx)-o.Radius,
		float64(my)-o.Radius,
	)
}

func loadImage(name string) *ebiten.Image {
	statikFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	file, err := statikFs.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	raw, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(raw)
}
