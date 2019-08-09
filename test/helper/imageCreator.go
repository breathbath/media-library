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

type ImageSpec struct {
	Format string
	Width int
	Height int
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
		imgSpec.Format = "jpg"
	}


	img := image.NewRGBA(image.Rect(0, 0, imgSpec.Width, imgSpec.Height))
	colors := make(map[int]color.RGBA, 2)

	colors[0] = color.RGBA{0, 100, 0, 255}   // green
	colors[1] = color.RGBA{50, 205, 50, 255} // limegreen

	index_color := 0
	size_board := 8
	size_block := int(imgSpec.Width / size_board)
	loc_x := 0

	for curr_x := 0; curr_x < size_board; curr_x++ {

		loc_y := 0
		for curr_y := 0; curr_y < size_board; curr_y++ {

			draw.Draw(img, image.Rect(loc_x, loc_y, loc_x+size_block, loc_y+size_block),
				&image.Uniform{colors[index_color]}, image.ZP, draw.Src)

			loc_y += size_block
			index_color = 1 - index_color // toggle from 0 to 1 to 0 to 1 to ...
		}
		loc_x += size_block
		index_color = 1 - index_color // toggle from 0 to 1 to 0 to 1 to ...
	}

	//imgRect := image.Rect(0, 0, imgSpec.Width, imgSpec.Height)
	//img := image.NewNRGBA(imgRect)
	//mygreen := color.RGBA{0, 100, 0, 255}
	//
	//draw.Draw(img, img.Bounds(), &image.Uniform{mygreen}, image.ZP, draw.Src)
	//for y := 0; y < imgSpec.Width; y += 10 {
	//	for x := 0; x < imgSpec.Height; x += 10 {
	//		fill := &image.Uniform{color.RGBA{11, 22, 32 ,255}}
	//		if rand.Intn(10)%2 == 0 {
	//			fill = &image.Uniform{color.RGBA{255, 255, 255 ,255}}
	//		}
	//		draw.Draw(img, image.Rect(x, y, x+10, y+10), fill, image.ZP, draw.Src)
	//	}
	//}

	buf := &bytes.Buffer{}

	var err error
	switch imgSpec.Format {
	case "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: imgSpec.Quality})
		break
	case "png":
		err = png.Encode(buf, img)
	default:
		return buf, fmt.Errorf("Unknown format '%s'", imgSpec.Format)
	}

	return buf, err
}
