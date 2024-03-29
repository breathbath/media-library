package assets

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"github.com/breathbath/go_utils/utils/env"
	io2 "github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-library/filesystem"
	"github.com/disintegration/imaging"
)

type ImageSaver struct {
	FileSystemHandler                      filesystem.Manager
	vertMaxImageWidth, horizMaxImageHeight int64
	jpegQuality                            int64
}

func NewImageSaver(fsHandler filesystem.Manager) ImageSaver {
	return ImageSaver{
		FileSystemHandler:   fsHandler,
		vertMaxImageWidth:   env.ReadEnvInt("VERT_MAX_IMAGE_WIDTH", 0),
		horizMaxImageHeight: env.ReadEnvInt("HORIZ_MAX_IMAGE_HEIGHT", 0),
		jpegQuality:         env.ReadEnvInt("COMPRESS_JPG_QUALITY", 85),
	}
}

func (is ImageSaver) SaveImage(sourceFile io.ReadSeeker, folderName, fileName string) error {
	io2.OutputInfo("", "Will save file %s in folder %s", fileName, folderName)
	targetFile, err := is.FileSystemHandler.CreateNonResizedFileWriter(folderName, fileName)
	defer func() {
		e := targetFile.Close()
		if e != nil {
			io2.OutputError(e, "", "Failed to close file '%s'", fileName)
		}
	}()

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
	sourceFile io.ReadSeeker,
	targetFile io.Writer,
	extWithDot string,
) error {
	ext := strings.TrimLeft(extWithDot, ".")
	_, err := sourceFile.Seek(0, 0)
	if err != nil {
		return err
	}

	imgRcr, _, err := image.Decode(sourceFile)
	if err != nil {
		return err
	}

	resizeX, resizeY := 0, 0
	if is.vertMaxImageWidth+is.horizMaxImageHeight > 0 {
		bounds := imgRcr.Bounds()
		if (bounds.Dy() > bounds.Dx() || bounds.Dy() == bounds.Dx()) && bounds.Dx() > int(is.vertMaxImageWidth) {
			resizeX = int(is.vertMaxImageWidth)
		} else if bounds.Dx() > bounds.Dy() && bounds.Dy() > int(is.horizMaxImageHeight) {
			resizeY = int(is.horizMaxImageHeight)
		}
	}

	if resizeX+resizeY > 0 {
		imgRcr = imaging.Resize(imgRcr, resizeX, resizeY, imaging.Lanczos)
	}

	if ext == "jpg" || ext == "jpeg" {
		return jpeg.Encode(targetFile, imgRcr, &jpeg.Options{Quality: int(is.jpegQuality)})
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

	io2.OutputWarning("", "Unknown file extension %s, will just copy file", ext)

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
