package cmds

import (
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/imgProcess"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
	"path"
)

// TestImageProcessing
// I want to detect if an image is sideways and needs to be rotated
// will be targeting specific examples for this
// Should only need to rotate 90 degrees for any sideways images
// I need to detect if the horizontal line is longer or shorter than the vertical line
// Horizontal line should be longer than the vertical line
func TestImageProcessing(commonDirs *config.CommonDirs) model.CliFunc {
	return func(cCtx *cli.Context) error {
		sideways1 := path.Join(commonDirs.ImageFolder,
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705",
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705-0.png")
		output1 := path.Join(commonDirs.TestRotation,
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705",
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705-0.png")

		processImage := imgProcess.NewProcessImage(sideways1, output1)
		return processImage.RotateImage()
	}
}
