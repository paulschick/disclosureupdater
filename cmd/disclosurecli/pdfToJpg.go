package main

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/urfave/cli/v2"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
)

// PdfToJpg converts PDF files to JPG files
// TODO - create struct for filepaths and fs operations
// - determine which files are already converted
// - get list of files to convert
// - create goroutines to convert files
func PdfToJpg(commonDirs *config.CommonDirs) CliFunc {
	return func(cCtx *cli.Context) error {
		pdfDir := commonDirs.DisclosuresFolder
		dirEntries, err := os.ReadDir(pdfDir)
		if err != nil {
			return err
		}

		for _, entry := range dirEntries {
			fmt.Printf("Processing %s\n", entry.Name())
			filePath := filepath.Join(pdfDir, entry.Name())
			fmt.Println(filePath)
			doc, err := fitz.New(filePath)
			if err != nil {
				return err
			}

			for n := 0; n < doc.NumPage(); n++ {
				img, err := doc.Image(n)
				if err != nil {
					return err
				}
				pdfName := entry.Name()
				baseFileName := strings.Split(pdfName, ".pdf")[0]
				imageDir := filepath.Join(commonDirs.ImageFolder, baseFileName)
				_, err = os.Stat(imageDir)
				imageDirExists := !os.IsNotExist(err)
				if imageDirExists {
					fmt.Printf("Image dir %s exists\n", imageDir)
					continue
				}

				err = util.TryCreateDirectories(imageDir)
				if err != nil {
					return err
				}
				imageName := fmt.Sprintf("%s-%d.jpg", baseFileName, n)
				f, err := os.Create(filepath.Join(imageDir, imageName))
				if err != nil {
					return err
				}
				err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
				if err != nil {
					return err
				}
				err = f.Close()
				if err != nil {
					return err
				}
			}

			_ = doc.Close()
		}
		return nil
	}
}
