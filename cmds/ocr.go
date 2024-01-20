package cmds

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/otiai10/gosseract/v2"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const MaxConversions = 25

func OcrImages(commonDirs *config.CommonDirs) model.CliFunc {
	return func(c *cli.Context) error {
		limit := c.Int("limit")
		if limit == 0 {
			limit = math.MaxInt
		}
		fmt.Printf("Limiting to %d\n", limit)

		imageDir := commonDirs.ImageFolder
		imageSubDirs, err := os.ReadDir(imageDir)
		if err != nil {
			return err
		}
		i := 0

		imagePaths := make([]string, 0)
		for _, imageSubDir := range imageSubDirs {
			if i >= limit {
				break
			}
			subDirPath := filepath.Join(imageDir, imageSubDir.Name())
			imageSubDirContents, err := os.ReadDir(subDirPath)
			if err != nil {
				return err
			}
			for _, imageFile := range imageSubDirContents {
				imagePath := filepath.Join(subDirPath, imageFile.Name())
				csvPath := filepath.Join(commonDirs.CsvFolder, csvPathFromImagePath(imagePath))
				_, err := os.Stat(csvPath)
				if err == nil {
					fmt.Printf("Skipping %s\n", imagePath)
					continue
				} else if !os.IsNotExist(err) {
					fmt.Printf("Error checking for csv: %s\n", err.Error())
					return err
				}
				imagePaths = append(imagePaths, imagePath)
				fmt.Printf("Adding %s\n", imagePath)
				i++
			}
		}

		waitChan := make(chan struct{}, MaxConversions)
		done := make(chan bool, len(imagePaths))
		failed := make(chan string, len(imagePaths))
		errs := make(chan error, len(imagePaths))

		index := 0
		for _, imgPath := range imagePaths {
			waitChan <- struct{}{}
			go func(imgPath string) {
				csvPath := filepath.Join(commonDirs.CsvFolder, csvPathFromImagePath(imgPath))
				created, err := extractImageToCsvIfNotExists(imgPath, csvPath)
				if err != nil {
					fmt.Printf("Error extracting image to csv: %s\n", err.Error())
					fmt.Printf("Failed Image Path: %s\n", imgPath)
					errs <- err
					done <- false
					failed <- imgPath
					return
				}
				if created {
					fmt.Printf("(%d) Created %s\n", index, csvPath)
					index++
				} else {
					fmt.Printf("(%d) Already Exists %s\n", index, csvPath)
					index++
				}
				errs <- nil
				done <- true
				failed <- ""
				<-waitChan
			}(imgPath)
		}
		failedImages := make([]string, 0)
		var errStr string
		for i := 0; i < len(imagePaths); i++ {
			if err := <-errs; err != nil {
				errStr = errStr + " " + err.Error()
			}
			if <-done == false {
				failedImages = append(failedImages, <-failed)
			} else {
				// drain failed channel
				<-failed
			}
		}

		failedPath := filepath.Join(commonDirs.CsvFolder, "failed.txt")
		failedFile, err := os.OpenFile(failedPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			return err
		}
		for _, failedImage := range failedImages {
			_, err = failedFile.WriteString(failedImage + "\n")
			if err != nil {
				return err
			}
		}

		err = failedFile.Close()
		if err != nil {
			return err
		}

		if errStr != "" {
			err = errors.New(errStr)
		}
		if err != nil {
			return err
		}
		return nil
	}
}

func csvPathFromImagePath(imagePath string) string {
	basePath := filepath.Base(imagePath)
	return strings.ReplaceAll(basePath, ".png", ".csv")
}

// extractImageToCsvIfNotExists returns true if the csv file was created, false if it already existed
// and an error if one occurred
func extractImageToCsvIfNotExists(imagePath, csvPath string) (bool, error) {
	var err error
	var client *gosseract.Client
	client, err = initializeClient(imagePath)
	if err != nil {
		return false, err
	}
	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = '\t'
		return gocsv.NewSafeCSVWriter(writer)
	})
	ocrResults := extractOcrResults(client)
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		csvFile, err := os.OpenFile(csvPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			return false, err
		}
		err = gocsv.MarshalFile(&ocrResults, csvFile)
		if err != nil {
			return false, err
		}
		err = csvFile.Close()
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

type OcrResult struct {
	LineNum    int     `csv:"lineNum"`
	WordNum    int     `csv:"wordNum"`
	Word       string  `csv:"word"`
	Confidence float64 `csv:"confidence"`
}

func NewOcrResult(box gosseract.BoundingBox) *OcrResult {
	return &OcrResult{
		LineNum:    box.LineNum,
		WordNum:    box.WordNum,
		Word:       strings.ReplaceAll(strings.ReplaceAll(box.Word, "\n", ""), " ", ""),
		Confidence: box.Confidence,
	}
}

func initializeClient(imagePath string) (*gosseract.Client, error) {
	var err error
	client := gosseract.NewClient()
	err = client.SetImage(imagePath)
	if err != nil {
		fmt.Printf("Error setting image: %s\n", err.Error())
		return nil, err
	}
	err = client.SetTessdataPrefix("/usr/share/tesseract-ocr/5/tessdata_best-4.1.0")
	if err != nil {
		fmt.Printf("Error setting tessdata prefix: %s\n", err.Error())
		return nil, err
	}
	err = client.SetLanguage("eng")
	if err != nil {
		fmt.Printf("Error setting language: %s\n", err.Error())
		return nil, err
	}
	return client, nil
}

func extractOcrResults(client *gosseract.Client) []*OcrResult {
	out, err := client.GetBoundingBoxesVerbose()
	if err != nil {
		fmt.Printf("Error getting bounding boxes: %s\n", err.Error())
		return nil
	}
	results := make([]*OcrResult, 0)
	for _, box := range out {
		result := NewOcrResult(box)
		results = append(results, result)
	}
	return results
}
