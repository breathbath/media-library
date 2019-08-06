package assets

import (
	"github.com/breathbath/go_utils/utils/env"
	"github.com/breathbath/media-service/fileSystem"
	"github.com/disintegration/imaging"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type ImageSaver struct {
	FileSystemHandler fileSystem.FileSystemManager
}

func (is ImageSaver) SaveImage(sourceFile multipart.File, folderName, fileName string) error {
	targetFile, err := is.FileSystemHandler.CreateNonResizedFileWriter(folderName, fileName)
	if err != nil {
		return err
	}

	err = is.SaveCompressedImageIfPossible(sourceFile, targetFile, filepath.Ext(fileName))
	if err != nil {
		return err
	}

	return nil
}

func (is ImageSaver) SaveCompressedImageIfPossible(
	sourceFile multipart.File,
	targetFile io.Writer,
	extWithDot string,
) error {
	ext := strings.TrimLeft(extWithDot, ".")
	vertMaxImageWidth := env.ReadEnvInt("VERT_MAX_IMAGE_WIDTH", 0)
	horizMaxImageHeight := env.ReadEnvInt("HORIZ_MAX_IMAGE_HEIGHT", 0)
	_, err := sourceFile.Seek(0, 0)
	if err != nil {
		return err
	}

	imgRcr, _, err := image.Decode(sourceFile)
	if err != nil {
		return err
	}

	resizeX, resizeY := 0, 0
	if vertMaxImageWidth+horizMaxImageHeight > 0 {
		bounds := imgRcr.Bounds()
		if (bounds.Dy() > bounds.Dx() || bounds.Dy() == bounds.Dx()) && bounds.Dx() > int(vertMaxImageWidth) {
			resizeX = int(vertMaxImageWidth)
		} else if bounds.Dx() > bounds.Dy() && bounds.Dy() > int(horizMaxImageHeight) {
			resizeY = int(horizMaxImageHeight)
		}
	}

	if resizeX+resizeY > 0 {
		imgRcr = imaging.Resize(imgRcr, resizeX, resizeY, imaging.Lanczos)
	}

	if ext == "jpg" || ext == "jpeg" {
		jpegQuality := env.ReadEnvInt("COMPRESS_JPG_QUALITY", 85)
		return jpeg.Encode(targetFile, imgRcr, &jpeg.Options{int(jpegQuality)})
	}

	if ext == "png" {
		encoder := &png.Encoder{
			CompressionLevel: png.BestSpeed,
		}
		return encoder.Encode(targetFile, imgRcr)
	}

	if ext == "gif" {
		return gif.Encode(targetFile, imgRcr, &gif.Options{})
	}

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
