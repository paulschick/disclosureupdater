package methods

import (
	"errors"
	"fmt"
	"os"
	"time"
)

func TryCreateDirectories(fp string) (err error) {
	if _, err = os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(fp, os.ModePerm)
	}
	return err
}

func CurrentYear() int {
	return time.Now().Year()
}

func GetCurrentYearString() string {
	return fmt.Sprintf("%d", CurrentYear())
}
