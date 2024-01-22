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
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const MaxConversions = 25

// OcrImages
// TODO 1. Reuse client, no need to re-create. Just SetImage for each
// TODO 2. Error Groups golang.org/x/sync/errgroup
// Instead of sending errors from goroutines to the main go routine
// TODO 3. Batch or Async IO operations - batch opening/closing files or async
// TODO 4. Buffered channels for done and failed chans
// TODO 5. Don't double-loop over the images
// TODO 6. Logging in a tight loop can be a performance hit
// Aggregate logs and use less frequently. Probably log per n number of files completed
// TODO 7. Parallel write files
func OcrImages(commonDirs *config.CommonDirs) model.CliFunc {
	return func(c *cli.Context) error {
		limit := c.Int("limit")
		if limit == 0 {
			limit = math.MaxInt
		}
		fmt.Printf("Limiting to %d\n", limit)
		imageDir := commonDirs.ImageFolder

		// ---- Example Implementation ----- //

		//tessClient, err := NewTessClientDefault()
		//if err != nil {
		//	fmt.Printf("Error creating new TessClient: %s\n", err.Error())
		//	return err
		//}
		//
		//imageIterator, err := NewImageIterator(imageDir)
		//if err != nil {
		//	fmt.Printf("Error creating new ImageIterator: %s\n", err.Error())
		//	return err
		//}
		//
		//
		//// Example Implementation
		//for imageIterator.HasNext() {
		//	var results []*model.OcrResult
		//	var extractor *ImageExtractor
		//	imagePath := imageIterator.GetNext()
		//
		//	if imagePath == "" {
		//		fmt.Printf("Error getting next image path.\nIteration complete\n")
		//		return nil
		//	}
		//
		//	extractor = NewImageExtractor(tessClient, imagePath, commonDirs)
		//	results, err = extractor.ExtractIfNotExists()
		//
		//	if err != nil {
		//		fmt.Printf("Error extracting image to csv: %s\n", err.Error())
		//		fmt.Printf("Failed Image Path: %s\n", imagePath)
		//		return err
		//	}
		//
		//	if results == nil {
		//		fmt.Printf("Skipping %s\n", imagePath)
		//		continue
		//	}
		//
		//	err = extractor.WriteResults(results)
		//	if err != nil {
		//		fmt.Printf("Error writing results: %s\n", err.Error())
		//		fmt.Printf("Failed Image Path: %s\n", imagePath)
		//		return err
		//	}
		//
		//	fmt.Printf("Created %s\n", extractor.CsvPath)
		//}

		// ---- End Example ----- //

		// ---- Original Working Code ----- //

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

func extractOcrResults(client *gosseract.Client) []*model.OcrResult {
	out, err := client.GetBoundingBoxesVerbose()
	if err != nil {
		fmt.Printf("Error getting bounding boxes: %s\n", err.Error())
		return nil
	}
	results := make([]*model.OcrResult, 0)
	for _, box := range out {
		result := model.NewOcrResult(box)
		results = append(results, result)
	}
	return results
}

type TessClient struct {
	TessDataPrefix string
	Language       string
	ImagePath      string
	Client         *gosseract.Client
}

func NewTessClientDefault() (*TessClient, error) {
	var err error
	lang := "eng"
	dataPrefix := "/usr/share/tesseract-ocr/5/tessdata_best-4.1.0"

	client := gosseract.NewClient()
	err = client.SetLanguage(lang)
	if err != nil {
		return nil, err
	}
	err = client.SetTessdataPrefix(dataPrefix)
	if err != nil {
		return nil, err
	}
	return &TessClient{
		TessDataPrefix: dataPrefix,
		Language:       lang,
		Client:         client,
	}, nil
}

func (t *TessClient) ExtractImageToResults(imagePath string) ([]*model.OcrResult, error) {
	var err error
	err = t.Client.SetImage(imagePath)
	if err != nil {
		return nil, err
	}
	out, err := t.Client.GetBoundingBoxesVerbose()
	if err != nil {
		return nil, err
	}
	results := make([]*model.OcrResult, len(out))
	for i, box := range out {
		results[i] = model.NewOcrResult(box)
	}
	return results, nil
}

type ImageExtractor struct {
	TessClient *TessClient
	ImagePath  string
	CsvPath    string
	CommonDirs *config.CommonDirs
}

func NewImageExtractor(tessClient *TessClient, imagePath string, commonDirs *config.CommonDirs) *ImageExtractor {
	return &ImageExtractor{
		TessClient: tessClient,
		ImagePath:  imagePath,
		CommonDirs: commonDirs,
		CsvPath:    filepath.Join(commonDirs.CsvFolder, csvPathFromImagePath(imagePath)),
	}
}

func (i *ImageExtractor) CsvExists() (bool, error) {
	_, err := os.Stat(i.CsvPath)
	if err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}

func (i *ImageExtractor) ExtractIfNotExists() ([]*model.OcrResult, error) {
	if exists, err := i.CsvExists(); err != nil {
		return nil, err
	} else if exists {
		return nil, nil
	}
	results, err := i.TessClient.ExtractImageToResults(i.ImagePath)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (i *ImageExtractor) WriteResults(results []*model.OcrResult) error {
	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = '\t'
		return gocsv.NewSafeCSVWriter(writer)
	})
	csvFile, err := os.OpenFile(i.CsvPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	err = gocsv.MarshalFile(results, csvFile)
	if err != nil {
		return err
	}
	err = csvFile.Close()
	if err != nil {
		return err
	}
	return nil
}

type NestedIterator interface {
	HasNext() bool
	GetNext() string
}

type ImageIterator struct {
	BaseDir       string
	SubDirs       []fs.DirEntry
	CurrentDir    string
	CurrentImages []fs.DirEntry
	CurrentIdx    int
	SubDirIdx     int
}

func NewImageIterator(baseDir string) (*ImageIterator, error) {
	subDirs, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}
	return &ImageIterator{
		BaseDir:    baseDir,
		SubDirs:    subDirs,
		CurrentIdx: -1,
		SubDirIdx:  0,
	}, nil
}

func (it *ImageIterator) HasNext() bool {
	for (it.CurrentIdx+1 >= len(it.CurrentImages) || len(it.CurrentImages) == 0) && it.SubDirIdx < len(it.SubDirs) {
		it.CurrentDir = filepath.Join(it.BaseDir, it.SubDirs[it.SubDirIdx].Name())
		var err error
		it.CurrentImages, err = os.ReadDir(it.CurrentDir)
		if err != nil {
			return false // Error handling can be more sophisticated here
		}
		it.CurrentIdx = -1
		it.SubDirIdx++
	}
	return it.CurrentIdx+1 < len(it.CurrentImages)
}

func (it *ImageIterator) GetNext() string {
	if it.HasNext() {
		it.CurrentIdx++
		return filepath.Join(it.CurrentDir, it.CurrentImages[it.CurrentIdx].Name())
	}
	return "" // or handle the "no more images" case more explicitly
}
