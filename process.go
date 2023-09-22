package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

func processImage(srcImage image.Image, config *Config) (image.Image, error) {
	scale := config.Scale
	borderWidth := config.BorderWidth

	tint := getAverageColorWithAdjustments(srcImage, 100, 10)

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
		return nil, errors.New("failed to convert scaledDstImage to *image.RGBA")
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
	dcScaled.SetColor(tint)
	dcScaled.DrawCircle(centerX, centerY, scaledRadius)
	dcScaled.Stroke()

	return dcScaled.Image(), nil
}

// Get the average color of an image with brightness and contrast adjustments
func getAverageColorWithAdjustments(img image.Image, brightnessAdjustment int, contrastAdjustment float64) color.RGBA {
	bounds := img.Bounds()
	var sumR, sumG, sumB uint64
	numPixels := uint64(bounds.Dx() * bounds.Dy())

	// Sum the color channels for all pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			sumR += uint64(r)
			sumG += uint64(g)
			sumB += uint64(b)
		}
	}

	// Calculate the mean values for each channel
	avgR := uint8(sumR / numPixels >> 8)
	avgG := uint8(sumG / numPixels >> 8)
	avgB := uint8(sumB / numPixels >> 8)

	// Apply brightness and contrast adjustments
	avgR = adjustBrightness(avgR, brightnessAdjustment)
	avgG = adjustBrightness(avgG, brightnessAdjustment)
	avgB = adjustBrightness(avgB, brightnessAdjustment)

	avgR = adjustContrast(avgR, avgR, contrastAdjustment)
	avgG = adjustContrast(avgG, avgG, contrastAdjustment)
	avgB = adjustContrast(avgB, avgB, contrastAdjustment)

	return color.RGBA{avgR, avgG, avgB, 255}
}

// Adjust brightness by adding a value to the color channel
func adjustBrightness(value uint8, adjustment int) uint8 {
	newValue := int(value) + adjustment
	if newValue > 255 {
		return 255
	} else if newValue < 0 {
		return 0
	}
	return uint8(newValue)
}

// Adjust contrast by scaling the difference from the mean color value
func adjustContrast(value, mean uint8, adjustment float64) uint8 {
	newValue := int(float64(value)-float64(mean)*adjustment) + int(mean)
	if newValue > 255 {
		return 255
	} else if newValue < 0 {
		return 0
	}
	return uint8(newValue)
}
