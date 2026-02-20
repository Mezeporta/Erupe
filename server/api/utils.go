package api

import (
	"errors"
	"path/filepath"

	"go.uber.org/zap"
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

func verifyPath(path string, trustedRoot string, logger *zap.Logger) (string, error) {

	c := filepath.Clean(path)
	logger.Debug("Cleaned path", zap.String("path", c))

	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		logger.Warn("Path verification failed", zap.Error(err))
		return c, errors.New("unsafe or invalid path specified")
	}

	err = inTrustedRoot(r, trustedRoot)
	if err != nil {
		logger.Warn("Path outside trusted root", zap.Error(err))
		return r, errors.New("unsafe or invalid path specified")
	} else {
		return r, nil
	}
}
