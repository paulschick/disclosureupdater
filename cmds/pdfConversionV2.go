package cmds

import (
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/paulschick/disclosureupdater/common/constants"
	"github.com/paulschick/disclosureupdater/common/logger"
	workerpool2 "github.com/paulschick/disclosureupdater/common/workerpool"
	"go.uber.org/zap"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
		zap.String("image_name", result.ImageName))

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
		zap.String("image_name", result.ImageName))
	return f.Close()
}

func getPdfEntries(slice bool, pdfDir string) ([]os.DirEntry, error) {
	pdfs, err := os.ReadDir(pdfDir)
	if err != nil {
		return nil, err
	}
	if slice {
		return pdfs[:200], nil
	}
	return pdfs, nil
}

func BatchPdfToPng(pdfDir, imageDir string) error {
	start := time.Now()
	logger.Logger.Info("Starting batch PDF to PNG conversion")

	// Rough calculation for workers based on number of CPUs
	// Tuned to one system currently
	numCpus := runtime.NumCPU()
	maxWorkers := int(math.Floor(float64(numCpus) * constants.CpuUtilization))

	// NOTE change to true for testing slices
	pdfs, err := getPdfEntries(false, pdfDir)
	if err != nil {
		return err
	}

	batches, err := calculateBatches(pdfs, pdfDir)
	if err != nil {
		return err
	}

	for i, batch := range batches {
		if i > 0 {
			batches[i-1] = nil
		}
		processBatch(batch, pdfDir, imageDir, maxWorkers)
	}

	elapsed := time.Since(start)
	logger.Logger.Info("Finished batch PDF to PNG conversion",
		zap.Duration("elapsed", elapsed))
	return nil
}

func processBatch(batch []os.DirEntry, pdfDir, imageDir string, poolSize int) {
	batchLen := len(batch)
	allTasks := make([]*workerpool2.Task, batchLen)
	for i := 0; i < batchLen; i++ {
		task := workerpool2.NewTask(func(data interface{}) error {
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
			data = nil
			return nil
		}, batch[i], i)
		allTasks[i] = task
	}

	pool := workerpool2.NewPool(allTasks, poolSize, batchLen)
	pool.Run()

	// Clear the slice for garbage collection
	for i := range allTasks {
		allTasks[i] = nil
	}
}

func calculateBatches(pdfs []os.DirEntry, pdfDir string) ([][]os.DirEntry, error) {
	var batches [][]os.DirEntry
	var currentBatch []os.DirEntry
	var currentPageCount int

	for _, pdf := range pdfs {
		pageCount, err := numberOfPagesInPdf(filepath.Join(pdfDir, pdf.Name()))
		if err != nil {
			return nil, err
		}
		if currentPageCount+pageCount > constants.BatchSize {
			batches = append(batches, currentBatch)
			currentBatch = nil
			currentPageCount = 0
		}

		currentBatch = append(currentBatch, pdf)
		currentPageCount += pageCount
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	return batches, nil
}

func numberOfPagesInPdf(pdfPath string) (int, error) {
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return 0, err
	}

	defer func() {
		closeErr := doc.Close()
		if closeErr != nil {
			logger.Logger.Error("Error closing pdf",
				zap.String("pdf_path", pdfPath),
				zap.Error(closeErr))
		}
	}()

	return doc.NumPage(), nil
}
