package assets

import (
	"fmt"
	"github.com/breathbath/go_utils/utils/io"
	"regexp"
	"strconv"
	"strings"
)

func isFolderValid(folderName string) bool {
	r := regexp.MustCompile(`^\w+$`)
	return r.MatchString(folderName)
}

func parseImageName(imageFile string) (imageName, imageExt string) {
	r := regexp.MustCompile(fmt.Sprintf(`^([\w\-_]+)\.(%s)$`, SUPPORTED_IMAGE_FORMATS))
	matches := r.FindStringSubmatch(imageFile)
	if len(matches) == 0 {
		return
	}

	imageName = matches[1]
	imageExt = matches[2]
	return
}

func extractSizes(path string) (width, height int) {
	width, height = 0, 0

	r := regexp.MustCompile(`(\d+x\d+)|(x\d+)|(\d+x)`)
	matches := r.FindStringSubmatch(path)
	if len(matches) == 0 {
		return
	}

	var e error
	if matches[1] != "" {
		r = regexp.MustCompile(`(\d+)x(\d+)`)
		matches2 := r.FindStringSubmatch(matches[1])
		width, e = strconv.Atoi(matches2[1])
		height, e = strconv.Atoi(matches2[2])
		if e != nil {
			io.OutputError(e, "", "")
		}
		return
	}

	if matches[2] != "" {
		r = regexp.MustCompile(`x(\d+)`)
		matches2 := r.FindStringSubmatch(matches[1])
		height, e = strconv.Atoi(matches2[1])
		if e != nil {
			io.OutputError(e, "", "")
		}
		return
	}

	r = regexp.MustCompile(`x(\d+)`)
	matches2 := r.FindStringSubmatch(matches[1])
	height, e = strconv.Atoi(matches2[1])
	if e != nil {
		io.OutputError(e, "", "")
	}
	return
}

func parseImagePath(path string) ImagePath {
	path = strings.Trim(path, "/")
	pathItems := strings.Split(path, "/")
	if len(pathItems) < 2 {
		return ImagePath{isValid: false}
	}

	if len(pathItems) == 2 {
		isValid := isFolderValid(pathItems[0])
		if !isValid {
			return ImagePath{isValid: false}
		}

		imageName, imageExt := parseImageName(pathItems[1])
		return ImagePath{
			folderName: pathItems[0],
			imageFile:  pathItems[1],
			imageName:  imageName,
			imageExt:   imageExt,
			isValid:    imageName != "" && imageExt != "",
		}
	}

	isValid := isFolderValid(pathItems[1])
	if !isValid {
		return ImagePath{isValid: false}
	}

	imageName, imageExt := parseImageName(pathItems[2])
	if imageName == "" || imageExt == "" {
		return ImagePath{isValid: false}
	}

	imagePath := ImagePath{
		folderName: pathItems[1],
		imageFile:  pathItems[2],
		imageName:  imageName,
		imageExt:   imageExt,
	}

	imagePath.width, imagePath.height = extractSizes(pathItems[0])
	if imagePath.width == 0 && imagePath.height == 0 {
		imagePath.isValid = false
		return imagePath
	}

	imagePath.rawResizedFolder = pathItems[0]
	imagePath.isValid = true

	return imagePath
}
