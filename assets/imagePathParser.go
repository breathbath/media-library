package assets

import (
	"fmt"
	"github.com/breathbath/go_utils/utils/io"
	"github.com/breathbath/media-library/fileSystem"
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
		matches2 := r.FindStringSubmatch(matches[2])
		height, e = strconv.Atoi(matches2[1])
		if e != nil {
			io.OutputError(e, "", "")
		}
		return
	}

	r = regexp.MustCompile(`(\d+)x`)
	matches2 := r.FindStringSubmatch(matches[3])
	width, e = strconv.Atoi(matches2[1])
	if e != nil {
		io.OutputError(e, "", "")
	}
	return
}

func parseImagePath(path string) fileSystem.ImagePath {
	path = strings.Trim(path, "/")
	pathItems := strings.Split(path, "/")
	if len(pathItems) < 2 {
		return fileSystem.ImagePath{IsValid: false}
	}

	if len(pathItems) == 2 {
		isValid := isFolderValid(pathItems[0])
		if !isValid {
			return fileSystem.ImagePath{IsValid: false}
		}

		imageName, imageExt := parseImageName(pathItems[1])
		return fileSystem.ImagePath{
			FolderName: pathItems[0],
			ImageFile:  pathItems[1],
			ImageName:  imageName,
			ImageExt:   imageExt,
			IsValid:    imageName != "" && imageExt != "",
		}
	}

	isValid := isFolderValid(pathItems[1])
	if !isValid {
		return fileSystem.ImagePath{IsValid: false}
	}

	imageName, imageExt := parseImageName(pathItems[2])
	if imageName == "" || imageExt == "" {
		return fileSystem.ImagePath{IsValid: false}
	}

	imagePath := fileSystem.ImagePath{
		FolderName: pathItems[1],
		ImageFile:  pathItems[2],
		ImageName:  imageName,
		ImageExt:   imageExt,
	}

	imagePath.Width, imagePath.Height = extractSizes(pathItems[0])
	if imagePath.Width == 0 && imagePath.Height == 0 {
		imagePath.IsValid = false
		return imagePath
	}

	imagePath.RawResizedFolder = pathItems[0]
	imagePath.IsValid = true

	return imagePath
}
