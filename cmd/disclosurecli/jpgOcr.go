package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

type ImageOcr struct {
	ImageFolder           string
	OcrFolder             string
	ImagePaths            []string
	ImageOcrOutputMapping map[string]string
}

func NewImageOcr(imageFolder, ocrFolder string) (*ImageOcr, error) {
	i := &ImageOcr{
		ImageFolder: imageFolder,
		OcrFolder:   ocrFolder,
	}
	err := i.SetImagePaths()
	if err != nil {
		return nil, err
	}
	i.ImageOcrOutputMapping = i.GetOcrOutputPaths()
	return i, nil
}

func (i *ImageOcr) SetImagePaths() error {
	imageEntries, err := os.ReadDir(i.ImageFolder)
	if err != nil {
		return err
	}
	for _, imageEntry := range imageEntries {
		imagePath := filepath.Join(i.ImageFolder, imageEntry.Name())
		i.ImagePaths = append(i.ImagePaths, imagePath)
	}
	return nil
}

func (i *ImageOcr) GetOcrOutputPaths() map[string]string {
	ocrOutputPaths := make(map[string]string)
	for _, imagePath := range i.ImagePaths {
		baseFileName := filepath.Base(imagePath)
		baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))] + ".txt"
		ocrOutputPath := filepath.Join(i.OcrFolder, baseFileName)
		ocrOutputPaths[imagePath] = ocrOutputPath
	}
	return ocrOutputPaths
}

func JpgOcr(commonDirs *config.CommonDirs) CliFunc {
	return func(cCtx *cli.Context) error {
		baseImageFolder := commonDirs.ImageFolder
		imageFolders, err := os.ReadDir(baseImageFolder)
		if err != nil {
			return err
		}
		imgOcrs := make([]*ImageOcr, 0)
		for _, imageFolder := range imageFolders {
			imgFolder := filepath.Join(baseImageFolder, imageFolder.Name())
			imgOcr, err := NewImageOcr(imgFolder, commonDirs.OcrFolder)
			if err != nil {
				return err
			}
			imgOcrs = append(imgOcrs, imgOcr)
		}
		imgOcrs = imgOcrs[:10]
		for _, imgOcr := range imgOcrs {
			fmt.Println(imgOcr.ImageOcrOutputMapping)
		}
		return nil
	}
}
