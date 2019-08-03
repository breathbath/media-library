package files

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
}
