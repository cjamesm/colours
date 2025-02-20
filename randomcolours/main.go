package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	blocks = 10
	blockw = 40
)

func main() {
	randomize := flag.Bool("randomize", true, "Randomize colors")
	flag.Parse()

	var c1, c2 colorful.Color
	var err error

	if *randomize {
		rand.Seed(time.Now().UnixNano())
		c1 = randomHexColor()
		c2 = randomHexColor()
	} else {
		c1, err = colorful.Hex("#fdffcc")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		c2, err = colorful.Hex("#242a42")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	img := createImage(blocks, blockw, 200)

	// Use these colors to get invalid RGB in the gradient.
	//c1, _ := colorful.Hex("#EEEF61")
	//c2, _ := colorful.Hex("#1E3140")

	// This can be used to "fix" invalid colors in the gradient.
	//draw.Draw(img, image.Rect(i*blockw,160,(i+1)*blockw,200), &image.Uniform{c1.BlendHcl(c2, float64(i)/float64(blocks-1)).Clamped()}, image.Point{}, draw.Src)

	drawIt(img, c1, c2)

	if err := saveImageWithUUID("colourblend", img); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func createImage(width, height, depth int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, width*blockw, depth))
}

func saveImage(filename string, img image.Image) error {
	toimg, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer toimg.Close()

	if err := png.Encode(toimg, img); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	return nil
}

func saveImageWithUUID(baseFilename string, img image.Image) error {
	uuid := uuid.New().String()[:5]
	filename := fmt.Sprintf("output/%s_%s.png", baseFilename, uuid)
	return saveImage(filename, img)
}

func drawIt(img *image.RGBA, c1, c2 colorful.Color) {
	for i := 0; i < blocks; i++ {
		draw.Draw(img, image.Rect(i*blockw, 0, (i+1)*blockw, 40), &image.Uniform{c1.BlendHsv(c2, float64(i)/float64(blocks-1))}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(i*blockw, 40, (i+1)*blockw, 80), &image.Uniform{c1.BlendLuv(c2, float64(i)/float64(blocks-1))}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(i*blockw, 80, (i+1)*blockw, 120), &image.Uniform{c1.BlendRgb(c2, float64(i)/float64(blocks-1))}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(i*blockw, 120, (i+1)*blockw, 160), &image.Uniform{c1.BlendLab(c2, float64(i)/float64(blocks-1))}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(i*blockw, 160, (i+1)*blockw, 200), &image.Uniform{c1.BlendHcl(c2, float64(i)/float64(blocks-1))}, image.Point{}, draw.Src)
	}
}

func randomHexColor() colorful.Color {
	return colorful.Color{
		R: rand.Float64(),
		G: rand.Float64(),
		B: rand.Float64(),
	}
}
