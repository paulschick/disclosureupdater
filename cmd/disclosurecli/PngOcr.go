package main

import (
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

type ImageOcr struct {
	ImageFolder           string
	OcrFolder             string
	OcrOutputFolder       string
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
	i.OcrOutputFolder = i.GetOcrOutputFolder()
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

func (i *ImageOcr) GetOcrOutputFolder() string {
	if len(i.ImagePaths) == 0 {
		fmt.Printf("No image paths found for %s\n", i.ImageFolder)
		return ""
	}
	imagePath := i.ImagePaths[0]
	imagePathFolder := filepath.Dir(imagePath)
	folderName := filepath.Base(imagePathFolder)
	return filepath.Join(i.OcrFolder, folderName)
}

func (i *ImageOcr) GetOcrOutputPaths() map[string]string {
	ocrOutputPaths := make(map[string]string)
	for _, imagePath := range i.ImagePaths {
		baseFileName := filepath.Base(imagePath)
		baseFileName = baseFileName[:len(baseFileName)-len(filepath.Ext(baseFileName))] + ".txt"
		ocrOutputPath := filepath.Join(i.OcrOutputFolder, baseFileName)
		ocrOutputPaths[imagePath] = ocrOutputPath
	}
	return ocrOutputPaths
}

func (i *ImageOcr) MakeOcrDir() error {
	return util.TryCreateDirectories(i.OcrOutputFolder)
}

func PngOcr(commonDirs *config.CommonDirs) CliFunc {
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
		//imgOcrs = imgOcrs[len(imgOcrs)-10:]
		imgOcrs = imgOcrs[:2]
		for _, imgOcr := range imgOcrs {
			outMap := imgOcr.ImageOcrOutputMapping
			outputPath := outMap[imgOcr.ImagePaths[0]]
			inputPath := imgOcr.ImagePaths[0]
			fmt.Printf("Input path: %s\n", inputPath)
			fmt.Printf("Output path: %s\n", outputPath)
			client := gosseract.NewClient()
			err = client.SetImage(inputPath)
			if err != nil {
				fmt.Printf("Unable to set image %s\n", inputPath)
				return err
			}
			text, _ := client.Text()
			fmt.Println(text)
			err = client.Close()
			if err != nil {
				fmt.Printf("Unable to close client for %s\n", inputPath)
				return err
			}
		}
		return nil
	}
}
