package fileSystem

import "path/filepath"

type ImagePath struct {//e.g. 350x360/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c.png
	FolderName       string //5d40a21343a5b
	ImageFile        string //801934c74f7fad05aa50a81bffa25e2c.png
	ImageName        string //801934c74f7fad05aa50a81bffa25e2c
	ImageExt         string //png
	RawResizedFolder string //350x360
	Width            int    //350
	Height           int    //360
	IsValid          bool   //if any part is invalid, the whole flag is false here
}

func (ip ImagePath) GetResizedFolderPath() string { // /var/www/static_public_html/cache/resized_image/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c
	return filepath.Join(
		"cache",
		"resized_image",
		ip.FolderName,
		ip.ImageName,
	)
}

func (ip ImagePath) GetResizedImagePath() string { // /var/www/static_public_html/cache/resized_image/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c/350x360.png
	resizedImage := ip.RawResizedFolder + "." + ip.ImageExt

	return filepath.Join(ip.GetResizedFolderPath(), resizedImage)
}

func (ip ImagePath) GetNonResizedImagePath() string { // /var/www/static_public_html/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c.png
	return filepath.Join(ip.GetNonResizedFolderPath(), ip.ImageFile)
}

func (ip ImagePath) GetNonResizedFolderPath() string { // /var/www/static_public_html/5d40a21343a5b
	return filepath.Join(ip.FolderName)
}