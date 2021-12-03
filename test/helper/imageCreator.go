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
)

const JPG = "jpg"

type ImageSpec struct {
	Format  string
	Width   int
	Height  int
	Quality int
}

func CreateImage(imgSpec ImageSpec) (io.Reader, error) {
	if imgSpec.Quality == 0 {
		imgSpec.Quality = 80
	}

	if imgSpec.Width == 0 {
		imgSpec.Width = 100
	}

	if imgSpec.Height == 0 {
		imgSpec.Height = 100
	}

	if imgSpec.Format == "" {
		imgSpec.Format = JPG
	}

	img := image.NewRGBA(image.Rect(0, 0, imgSpec.Width, imgSpec.Height))
	colors := make(map[int]color.RGBA, 2)

	colors[0] = color.RGBA{0, 100, 0, 255}   // green
	colors[1] = color.RGBA{50, 205, 50, 255} // limegreen

	indexColor := 0
	sizeBoard := 8
	sizeBlock := imgSpec.Width / sizeBoard
	locX := 0

	for currX := 0; currX < sizeBoard; currX++ {
		locY := 0
		for currY := 0; currY < sizeBoard; currY++ {
			draw.Draw(img, image.Rect(locX, locY, locX+sizeBlock, locY+sizeBlock),
				&image.Uniform{colors[indexColor]}, image.Point{}, draw.Src)

			locY += sizeBlock
			indexColor = 1 - indexColor // toggle from 0 to 1 to 0 to 1 to ...
		}
		locX += sizeBlock
		indexColor = 1 - indexColor // toggle from 0 to 1 to 0 to 1 to ...
	}

	buf := &bytes.Buffer{}

	var err error
	switch imgSpec.Format {
	case "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: imgSpec.Quality})
	case "png":
		err = png.Encode(buf, img)
	default:
		return buf, fmt.Errorf("unknown format '%s'", imgSpec.Format)
	}

	return buf, err
}
