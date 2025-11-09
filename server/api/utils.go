package api

import (
	"errors"
	"fmt"
	"path/filepath"
)

func inTrustedRoot(path string, trustedRoot string) error {
	for path != "/" {
		path = filepath.Dir(path)
		if path == trustedRoot {
			return nil
		}
	}
	return errors.New("path is outside of trusted root")
}

func verifyPath(path string, trustedRoot string) (string, error) {

	c := filepath.Clean(path)
	fmt.Println("Cleaned path: " + c)

	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		fmt.Println("Error " + err.Error())
		return c, errors.New("unsafe or invalid path specified")
	}

	err = inTrustedRoot(r, trustedRoot)
	if err != nil {
		fmt.Println("Error " + err.Error())
		return r, errors.New("unsafe or invalid path specified")
	} else {
		return r, nil
	}
}
