package util

import (
	"errors"
	"github.com/paulschick/disclosureupdater/model"
	"os"
	"path"
)

func TryCreateDirectories(fp string) (err error) {
	if _, err = os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(fp, os.ModePerm)
	}
	return err
}

func CreatePdfDownloadDirectory(dataFolder string) error {
	return TryCreateDirectories(path.Join(dataFolder, model.BasePdfDir))
}
