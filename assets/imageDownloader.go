package assets

import (
	io2 "github.com/breathbath/go_utils/utils/io"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadFile(url, originalPath string) (http.File, error) {
	var AppFs = afero.NewOsFs()
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(){
		err := resp.Body.Close()
		if err != nil {
			io2.OutputError(err, "", "")
		}
	}()

	if resp.StatusCode < 199 || resp.StatusCode > 299 {
		return nil, &os.PathError{Op: "open", Path: originalPath, Err: os.ErrNotExist}
	}

	err = AppFs.MkdirAll("/tmp", os.ModePerm)
	if err != nil {
		return nil, err
	}
	
	// Create the file
	targetFile, err := AppFs.Create(filepath.Join("/tmp", originalPath))
	if err != nil {
		return nil, err
	}

	// Write the body to file
	_, err = io.Copy(targetFile, resp.Body)
	if err != nil {
		return nil, err
	}

	_, err = targetFile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	return targetFile, err
}
