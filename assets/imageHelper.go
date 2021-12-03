package assets

import (
	"path/filepath"
	"regexp"
)

func SanitizeImageName(fullName string) string {
	ext := filepath.Ext(fullName)
	imageName := fullName[0 : len(fullName)-len(ext)]
	r := regexp.MustCompile(`[^\w\-]`)
	sanitizedImageName := r.ReplaceAllString(imageName, "_")

	return sanitizedImageName + ext
}
