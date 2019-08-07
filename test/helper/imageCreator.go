package helper

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
)

func CreateImage(format string) (io.Reader, error) {
	// generate some QR code look a like image

	imgRect := image.Rect(0, 0, 100, 100)
	img := image.NewGray(imgRect)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
	for y := 0; y < 100; y += 10 {
		for x := 0; x < 100; x += 10 {
			fill := &image.Uniform{color.Black}
			if rand.Intn(10)%2 == 0 {
				fill = &image.Uniform{color.White}
			}
			draw.Draw(img, image.Rect(x, y, x+10, y+10), fill, image.ZP, draw.Src)
		}
	}

	buf := &bytes.Buffer{}

	var err error
	switch format {
	case "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 80})
		break
	case "png":
		err = png.Encode(buf, img)
	default:
		return buf, fmt.Errorf("Unknown format '%s'", format)
	}

	return buf, err
}
