package files

import "path/filepath"

type ImagePath struct {
	//e.g. 350x360/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c.png
	folderName       string //5d40a21343a5b
	imageFile        string //801934c74f7fad05aa50a81bffa25e2c.png
	imageName        string //801934c74f7fad05aa50a81bffa25e2c
	imageExt         string //png
	rawResizedFolder string //350x360
	width            int    //350
	height           int    //360
	isValid          bool   //if any part is invalid, the whole flag is false here
	assetsPath       string //e.g. /var/www/static_public_html
}

func (ip ImagePath) GetResizedFolderPath() string { // /var/www/static_public_html/cache/resized_image/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c
	return filepath.Join(
		ip.assetsPath,
		"cache",
		"resized_image",
		ip.folderName,
		ip.imageName,
	)
}

func (ip ImagePath) GetResizedImagePath() string { // /var/www/static_public_html/cache/resized_image/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c/350x360.png
	resizedImage := ip.rawResizedFolder + "." + ip.imageExt

	return filepath.Join(ip.GetResizedFolderPath(), resizedImage)
}

func (ip ImagePath) GetNonResizedImagePath() string { // /var/www/static_public_html/5d40a21343a5b/801934c74f7fad05aa50a81bffa25e2c.png
	return filepath.Join(ip.GetNonResizedFolderPath(), ip.imageFile)
}

func (ip ImagePath) GetNonResizedFolderPath() string { // /var/www/static_public_html/5d40a21343a5b
	return filepath.Join(ip.assetsPath, ip.folderName)
}