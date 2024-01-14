package main

import (
	"errors"
	"fmt"
	"github.com/gen2brain/go-fitz"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/urfave/cli/v2"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

const MaxJobs = 25

type PdfConverter struct {
	PdfPath      string
	CommonDirs   *config.CommonDirs
	BaseFileName string
	ImageDir     string
}

func NewPdfConverter(pdfPath string, commonDirs *config.CommonDirs) *PdfConverter {
	p := &PdfConverter{
		PdfPath:    pdfPath,
		CommonDirs: commonDirs,
	}
	p.setBaseFileName()
	p.setImageDir()
	return p
}

func (p *PdfConverter) setBaseFileName() {
	baseFilePath := strings.Split(p.PdfPath, ".pdf")[0]
	p.BaseFileName = filepath.Base(baseFilePath)
}

func (p *PdfConverter) setImageDir() {
	fmt.Println(p.BaseFileName)
	p.ImageDir = filepath.Join(p.CommonDirs.ImageFolder, p.BaseFileName)
}

func (p *PdfConverter) ImageDirExists() (bool, error) {
	fmt.Printf("Checking if image dir %s exists\n", p.ImageDir)
	_, err := os.Stat(p.ImageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *PdfConverter) ConvertIfNotPresent() error {
	exists, err := p.ImageDirExists()
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Image dir %s exists\n", p.ImageDir)
		return nil
	}
	err = p.CreateImageDir()
	if err != nil {
		return err
	}
	fmt.Printf("Created dir %s\n", p.ImageDir)
	doc, err := fitz.New(p.PdfPath)
	if err != nil {
		return err
	}

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			return err
		}
		imageName := p.GetImageName(n)
		f, err := os.Create(filepath.Join(p.ImageDir, imageName))
		if err != nil {
			return err
		}
		fmt.Printf("processing image %s\n", imageName)
		err = png.Encode(f, img)
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			return err
		}
		fmt.Printf("created image %s\n", imageName)
	}
	err = doc.Close()
	if err != nil {
		return err
	}
	fmt.Printf("Converted %s\n", p.PdfPath)
	return nil
}

func (p *PdfConverter) CreateImageDir() error {
	return util.TryCreateDirectories(p.ImageDir)
}

func (p *PdfConverter) GetImageName(pageNumber int) string {
	return fmt.Sprintf("%s-%d.png", p.BaseFileName, pageNumber)
}

func (p *PdfConverter) CreateImageFile(pageNumber int) (*os.File, error) {
	imageName := p.GetImageName(pageNumber)
	return os.Create(filepath.Join(p.ImageDir, imageName))
}

// PdfToPng converts PDF files to PNG files
func PdfToPng(commonDirs *config.CommonDirs) CliFunc {
	return func(cCtx *cli.Context) error {
		pdfDir := commonDirs.DisclosuresFolder
		dirEntries, err := os.ReadDir(pdfDir)
		if err != nil {
			return err
		}

		pdfConverters := make([]*PdfConverter, 0)

		for _, entry := range dirEntries {
			filePath := filepath.Join(pdfDir, entry.Name())
			pdfConverter := NewPdfConverter(filePath, commonDirs)
			pdfConverters = append(pdfConverters, pdfConverter)
		}

		waitChan := make(chan struct{}, MaxJobs)
		count := 0
		done := make(chan bool, len(pdfConverters))
		errs := make(chan error, len(pdfConverters))
		for _, pdfConverter := range pdfConverters {
			waitChan <- struct{}{}
			count++
			go func(pdfConverter *PdfConverter) {
				err := pdfConverter.ConvertIfNotPresent()
				if err != nil {
					errs <- err
					done <- false
					return
				}
				done <- true
				errs <- nil
				<-waitChan
			}(pdfConverter)
		}
		var errStr string
		for i := 0; i < len(pdfConverters); i++ {
			if err := <-errs; err != nil {
				errStr = errStr + " " + err.Error()
			}
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
