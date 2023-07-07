package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

func ProcessImage(imagePath string, c color.Color) (string, error) {
	scale := 300.0
	borderWidth := 10.0
	// Placeholder: Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	srcImage, _, err := image.Decode(file)
	if err != nil {
		return "", err
	}

	// Create a new transparent image with the same dimensions as the source image
	dstImage := image.NewRGBA(srcImage.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Create a context for drawing on the destination image
	dc := gg.NewContextForRGBA(dstImage)
	dc.SetRGB(1, 1, 1) // Set the background color (white)

	// Calculate the center and radius of the circle
	width := float64(dstImage.Bounds().Dx())
	height := float64(dstImage.Bounds().Dy())
	radius := math.Min(width, height) / 2

	// Draw a circle on the destination image
	dc.DrawCircle(width/2, height/2, radius)
	dc.Clip()
	dc.Fill()

	// Draw the source image onto the destination image, cropped to the circle
	dc.DrawImage(srcImage, 0, 0)

	// Scale down the destination image to 300x300
	scaledDstImage := resize.Resize(uint(scale), uint(scale), dstImage, resize.Lanczos3)

	scaledDstRGBA, ok := scaledDstImage.(*image.RGBA)

	if !ok {
		return "", errors.New("failed to convert scaledDstImage to *image.RGBA")
	}

	// Create a new context for drawing on the scaled destination image
	dcScaled := gg.NewContextForRGBA(scaledDstRGBA)

	// Calculate the center and radius of the scaled circle
	scaledWidth := scale
	scaledHeight := scale
	scaledRadius := (math.Min(scaledWidth-borderWidth, scaledHeight-borderWidth) / 2)

	// Set the circle center and radius
	centerX := scaledWidth / 2
	centerY := scaledHeight / 2

	// Draw a circle border on the scaled destination image
	dcScaled.SetLineWidth(borderWidth)
	dcScaled.SetColor(c)
	dcScaled.DrawCircle(centerX, centerY, scaledRadius)
	dcScaled.Stroke()

	newFilename := "token_" + filepath.Base(strings.Replace(imagePath, "danimalsound_", "", 1))
	newFilePath := filepath.Join(filepath.Dir(imagePath), newFilename)
	newFile, err := os.Create(newFilePath)

	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	err = png.Encode(newFile, dcScaled.Image())
	if err != nil {
		log.Fatal(err)
	}

	return newFilePath, nil
}

// parseHexColor parses a hex color code (#RRGGBB) and returns the corresponding color.Color.
func parseHexColor(hexCode string) color.Color {
	hexCode = strings.TrimPrefix(hexCode, "#")

	if len(hexCode) != 6 {
		return color.Black
	}

	r := parseHexComponent(hexCode[0:2])
	g := parseHexComponent(hexCode[2:4])
	b := parseHexComponent(hexCode[4:6])

	return color.RGBA{r, g, b, 255}
}

// parseHexComponent parses a hexadecimal component and returns its integer value.
func parseHexComponent(hexComponent string) uint8 {
	var value uint64
	fmt.Sscanf(hexComponent, "%02x", &value)
	return uint8(value)
}
