package assets

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"

	io2 "github.com/breathbath/go_utils/utils/io"
	"github.com/spf13/afero"
)

func DownloadFile(url, originalPath string) (http.File, error) {
	var AppFs = afero.NewOsFs()
	resp, err := http.Get(url) // nolint: gosec, noctx
	if err != nil {
		return nil, err
	}

	defer func() {
		e := resp.Body.Close()
		if e != nil {
			io2.OutputError(e, "", "")
		}
	}()

	if resp.StatusCode < 199 || resp.StatusCode > 299 {
		var respBytes []byte
		respBytes, err = httputil.DumpResponse(resp, true)
		if err != nil {
			io2.OutputError(err, "", "")
		} else {
			io2.OutputInfo("", "Proxy source returned response: %s", string(respBytes))
		}
		return nil, &os.PathError{Op: "open", Path: originalPath, Err: os.ErrNotExist}
	}

	err = AppFs.MkdirAll("/tmp", os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create the file
	targetFile, err := AppFs.Create("/" + filepath.Join("tmp", originalPath))
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
