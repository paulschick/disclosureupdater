package cmds

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func CleanupImages(commonDirs *config.CommonDirs) model.CliFunc {
	return func(cCtx *cli.Context) error {
		removeCount := 0
		var err error
		imageDir := commonDirs.ImageFolder
		imageSubDirs, err := os.ReadDir(imageDir)
		if err != nil {
			return err
		}
		for _, imageSubDir := range imageSubDirs {
			subDirPath := filepath.Join(imageDir, imageSubDir.Name())
			imageSubDirContents, err := os.ReadDir(subDirPath)
			if err != nil {
				return err
			}
			if len(imageSubDirContents) == 0 {
				fmt.Println("Removing empty directory: ", imageSubDir.Name())
				err = os.Remove(imageSubDir.Name())
				if err != nil {
					return err
				}
				removeCount++
			}
		}
		fmt.Println("Removed ", removeCount, " empty directories")
		return err
	}
}
