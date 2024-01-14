package main

import (
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/imgProcess"
	"github.com/urfave/cli/v2"
	"image"
	"path"
)

// TestImageProcessing
// I want to detect if an image is sideways and needs to be rotated
// will be targeting specific examples for this
// Should only need to rotate 90 degrees for any sideways images
// I need to detect if the horizontal line is longer or shorter than the vertical line
// Horizontal line should be longer than the vertical line
func TestImageProcessing(commonDirs *config.CommonDirs) CliFunc {
	return func(cCtx *cli.Context) error {
		sideways1 := path.Join(commonDirs.ImageFolder,
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705",
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705-0.png")
		output1 := path.Join(commonDirs.TestRotation,
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705",
			"2013.ptr-pdfs.FL09.Grayson.Alan.9105705-0.png")

		processImage := imgProcess.NewProcessImage(sideways1, output1)
		img, err := processImage.OpenImage()
		if err != nil {
			return err
		}
		tensor := processImage.ImageToTensor(img)
		rotatedTensor := processImage.RotateImage90DegreesLeft(tensor)
		newImage, err := processImage.TensorToImage(rotatedTensor)
		if err != nil {
			return err
		}
		err = processImage.SaveImage(newImage.(*image.RGBA))
		if err != nil {
			return err
		}
		return nil
	}
}
