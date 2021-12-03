package filesystem

import "path/filepath"

type ImagePath struct {
	FolderName       string
	ImageFile        string
	ImageName        string
	ImageExt         string
	RawResizedFolder string
	Width            int
	Height           int
	IsValid          bool
}

func (ip *ImagePath) GetResizedFolderPath() string {
	return filepath.Join(
		"cache",
		"resized_image",
		ip.FolderName,
		ip.ImageName,
	)
}

func (ip *ImagePath) GetResizedParentFolderPath() string {
	return filepath.Join(
		"cache",
		"resized_image",
		ip.FolderName,
	)
}

func (ip *ImagePath) GetResizedImagePath() string {
	resizedImage := ip.RawResizedFolder + "." + ip.ImageExt

	return filepath.Join(ip.GetResizedFolderPath(), resizedImage)
}

func (ip *ImagePath) GetNonResizedImagePath() string {
	return filepath.Join(ip.GetNonResizedFolderPath(), ip.ImageFile)
}

func (ip *ImagePath) GetNonResizedFolderPath() string {
	return filepath.Join(ip.FolderName)
}
