package cmds

import (
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
)

func UpdateFolders(commonDirs *config.CommonDirs) model.CliFunc {
	return func(c *cli.Context) error {
		newImages := c.String("images")
		newDisclosures := c.String("disclosures")
		newOcr := c.String("ocr")
		newCsv := c.String("csv")

		// TODO - need to update config package
		// Set to current values
		if newImages == "" {
			newImages = commonDirs.ImageFolder
		}
		if newDisclosures == "" {
			newDisclosures = commonDirs.DisclosuresFolder
		}
		if newOcr == "" {
			newOcr = commonDirs.OcrFolder
		}
		if newCsv == "" {
			newCsv = commonDirs.CsvFolder
		}

		return nil
	}
}
