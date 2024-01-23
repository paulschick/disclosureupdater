package cmds

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/paulschick/disclosureupdater/logger"
	"go.uber.org/zap"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type ConversionResult struct {
	Image     image.Image
	ImageName string
	ImageDir  string
}

type PdfConverterV2 struct {
	PdfPath      string
	BaseFileName string
	ImageDir     string
}

// NewPdfConverterV2
// TODO - use a flat file structure instead of the nested structure
// images will be organized by {name}-{page no}.png, so they'll maintain order
func NewPdfConverterV2(pdfPath, imageDir string) *PdfConverterV2 {
	return &PdfConverterV2{
		PdfPath:      pdfPath,
		ImageDir:     imageDir,
		BaseFileName: filepath.Base(strings.Split(pdfPath, ".pdf")[0]),
	}
}

func (p *PdfConverterV2) GetImageName(pageNumber int, extension string) string {
	return fmt.Sprintf("%s-%d%s", p.BaseFileName, pageNumber, extension)
}

func (p *PdfConverterV2) ConvertPagesToImages(extension string) ([]ConversionResult, error) {
	var results []ConversionResult
	var err error
	var doc *fitz.Document

	doc, err = fitz.New(p.PdfPath)
	if err != nil {
		return nil, err
	}

	defer func(err error) {
		err = doc.Close()
	}(err)

	for n := 0; n < doc.NumPage(); n++ {
		var img image.Image
		img, err = doc.Image(n)
		if err != nil {
			return nil, err
		}
		imageName := p.GetImageName(n, extension)
		results = append(results, ConversionResult{
			Image:     img,
			ImageName: imageName,
			ImageDir:  p.ImageDir,
		})
	}

	return results, err
}

func WriteImage(result ConversionResult) error {
	logger.Logger.Info("Writing image",
		zap.String("image_name", result.ImageName),
		zap.String("image_dir", result.ImageDir))

	var err error
	var f *os.File

	f, err = os.Create(filepath.Join(result.ImageDir, result.ImageName))

	if err != nil {
		_ = f.Close()
		return err
	}

	err = png.Encode(f, result.Image)

	if err != nil {
		_ = f.Close()
		return err
	}

	logger.Logger.Info("Finished writing image",
		zap.String("image_name", result.ImageName),
		zap.String("image_dir", result.ImageDir))
	return f.Close()
}

func BatchPdfToPng(pdfDir, imageDir string) error {
	logger.Logger.Info("Starting batch PDF to PNG conversion")
	pdfs, err := os.ReadDir(pdfDir)
	if err != nil {
		return err
	}
	pdfs = pdfs[:100]

	numCpu := runtime.NumCPU()
	maxJobs := numCpu * 2
	waitChan := make(chan struct{}, maxJobs)
	var wg sync.WaitGroup
	var writeWg sync.WaitGroup

	for _, pdf := range pdfs {
		waitChan <- struct{}{}
		wg.Add(1)
		go func(pdf os.DirEntry) {
			defer wg.Done()

			pdfConverter := NewPdfConverterV2(filepath.Join(pdfDir, pdf.Name()), imageDir)
			results, err := pdfConverter.ConvertPagesToImages(".png")
			if err != nil {
				// TODO - handle error
				fmt.Println(err)
			}

			for _, result := range results {
				writeWg.Add(1)
				go func(res ConversionResult) {
					defer writeWg.Done()
					if err = WriteImage(res); err != nil {
						// TODO - handle error
						fmt.Printf("error writing image %s\n", res.ImageName)
						fmt.Println(err)
					}
				}(result)
			}

			<-waitChan
		}(pdf)
	}

	wg.Wait()
	writeWg.Wait()
	logger.Logger.Info("Finished batch PDF to PNG conversion")
	return nil
}
