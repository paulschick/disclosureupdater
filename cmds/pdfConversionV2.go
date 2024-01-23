package cmds

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/paulschick/disclosureupdater/logger"
	"github.com/paulschick/disclosureupdater/workerpool"
	"go.uber.org/zap"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	MaxWorkers = 12
	BatchSize  = 250
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
		img = nil
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

	result.Image = nil

	logger.Logger.Info("Finished writing image",
		zap.String("image_name", result.ImageName),
		zap.String("image_dir", result.ImageDir))
	return f.Close()
}

func getPdfEntries(slice bool, pdfDir string) ([]os.DirEntry, error) {
	pdfs, err := os.ReadDir(pdfDir)
	if err != nil {
		return nil, err
	}
	if slice {
		return pdfs[:500], nil
	}
	return pdfs, nil
}

func BatchPdfToPng(pdfDir, imageDir string) error {
	start := time.Now()
	logger.Logger.Info("Starting batch PDF to PNG conversion")

	pdfs, err := getPdfEntries(true, pdfDir)
	if err != nil {
		return err
	}
	totalPdfs := len(pdfs)

	for startIdx := 0; startIdx < totalPdfs; startIdx += BatchSize {
		end := startIdx + BatchSize
		if end > totalPdfs {
			end = totalPdfs
		}
		batch := pdfs[startIdx:end]
		processBatch(batch, pdfDir, imageDir, MaxWorkers)
	}

	elapsed := time.Since(start)
	logger.Logger.Info("Finished batch PDF to PNG conversion",
		zap.Duration("elapsed", elapsed))
	return nil
}

func processBatch(batch []os.DirEntry, pdfDir, imageDir string, poolSize int) {
	batchLen := len(batch)
	allTasks := make([]*workerpool.Task, batchLen)
	for i := 0; i < batchLen; i++ {
		task := workerpool.NewTask(func(data interface{}) error {
			entry := data.(os.DirEntry)
			pdfConverter := NewPdfConverterV2(filepath.Join(pdfDir, entry.Name()), imageDir)
			results, err := pdfConverter.ConvertPagesToImages(".png")
			pdfConverter = nil
			if err != nil {
				logger.Logger.Error("error converting pdf to images", zap.Error(err))
				return err
			}
			// TODO - Create a new task for each image
			for _, result := range results {
				if err = WriteImage(result); err != nil {
					logger.Logger.Error("error writing image",
						zap.String("image_name", result.ImageName),
						zap.String("image_dir", result.ImageDir),
						zap.Error(err))
					return err
				}
			}
			logger.Logger.Info("Finished batch PDF to PNG conversion",
				zap.Int("Task ID", i),
				zap.String("pdf_name", entry.Name()))
			return nil
		}, batch[i], i)
		allTasks[i] = task
	}

	pool := workerpool.NewPool(allTasks, poolSize, batchLen)
	pool.Run()

	// Clear the slice for garbage collection
	for i := range allTasks {
		allTasks[i] = nil
	}
}
