package downloader

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/paulschick/disclosureupdater/common/paths"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/paulschick/disclosureupdater/util"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GenerateZipUrlForYear(year int) string {
	return strings.Replace(util.ZipUrlTemplate, "{YEAR}", strconv.Itoa(year), 1)
}

func GenerateAllZipUrls() []string {
	minYear := util.MinYear
	year := util.CurrentYear()
	numYears := year - minYear + 1

	downloadUrls := make([]string, numYears)

	for i := minYear; i <= year; i++ {
		downloadUrls[i-minYear] = GenerateZipUrlForYear(i)
	}
	return downloadUrls
}

type DisclosureDownload struct {
	Url          string
	FileName     string
	BaseFilePath string
	ZipPath      string
	XmlPath      string
	CsvPath      string
}

func NewDisclosureDownload(url, baseFolder string) *DisclosureDownload {
	fileName, fileExt := paths.FileAndExtension(url)
	zipPath := path.Join(baseFolder, fmt.Sprintf("%s%s", fileName, fileExt))
	xmlPath := path.Join(baseFolder, fmt.Sprintf("%s.xml", fileName))
	csvPath := path.Join(baseFolder, fmt.Sprintf("%s.csv", fileName))
	tempFilePath := path.Join(os.TempDir(), fmt.Sprintf("%s%s", fileName, fileExt))
	return &DisclosureDownload{
		Url:          url,
		FileName:     fileName,
		BaseFilePath: tempFilePath,
		ZipPath:      zipPath,
		XmlPath:      xmlPath,
		CsvPath:      csvPath,
	}
}

func (d *DisclosureDownload) ToString() string {
	return fmt.Sprintf("Url: %s\nZip Path: %s\nXML Path: %s\nCSV Path: %s\n", d.Url, d.ZipPath, d.XmlPath, d.CsvPath)
}

type Downloadable struct {
	Url   string
	Bytes []byte
	Fp    string
}

func DownloadFile(url, outFilePath string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return err
	}
	file, err := os.Create(outFilePath)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func (d *DisclosureDownload) Download() error {
	return DownloadFile(d.Url, d.ZipPath)
}

func (d *DisclosureDownload) ZipIsPresent() bool {
	_, err := os.Stat(d.ZipPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (d *DisclosureDownload) XmlIsPresent() bool {
	_, err := os.Stat(d.XmlPath)
	return !errors.Is(err, os.ErrNotExist)
}

func (d *DisclosureDownload) Extract() error {
	arch, err := zip.OpenReader(d.ZipPath)
	if err != nil {
		return err
	}

	if len(arch.File) != 2 {
		fmt.Printf("Expected 2 files in zip, found %d\n", len(arch.File))
		return nil
	}

	var f *zip.File

	for _, file := range arch.File {
		if strings.HasSuffix(file.Name, ".xml") {
			f = file
			break
		}
	}

	if f == nil {
		fmt.Printf("No XML file found in zip\n")
		return nil
	}

	filePath := d.XmlPath
	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	fArchive, err := f.Open()
	if err != nil {
		return err
	}
	if _, err := io.Copy(destFile, fArchive); err != nil {
		return err
	}
	err = fArchive.Close()
	if err != nil {
		return err
	}
	err = destFile.Close()
	if err != nil {
		return err
	}
	err = arch.Close()
	if err != nil {
		return err
	}
	err = os.Remove(d.ZipPath)
	if err != nil {
		return err
	}

	return nil
}

func DownloadZipsIfNotPresent(downloads []*DisclosureDownload) error {
	var err error
	var wg sync.WaitGroup
	wg.Add(len(downloads))
	for _, disclosureDownload := range downloads {
		go func(d *DisclosureDownload, wg *sync.WaitGroup) {
			defer wg.Done()
			if !d.XmlIsPresent() {
				if !d.ZipIsPresent() {
					fmt.Printf("Downloading %s\n", d.ZipPath)
					err = d.Download()
					if err != nil {
						log.Fatalf("Error downloading %s: %s", d.Url, err)
					}
					err = d.Extract()
					if err != nil {
						log.Fatalf("Error extracting %s: %s", d.ZipPath, err)
					}
				} else {
					fmt.Printf("Skipping download of %s\n", d.ZipPath)
					err = d.Extract()
					if err != nil {
						log.Fatalf("Error extracting %s: %s", d.ZipPath, err)
					}
				}
			} else {
				fmt.Printf("Skipping download of %s\n", d.XmlPath)
			}
		}(disclosureDownload, &wg)
	}
	wg.Wait()
	return err
}

func (d *Downloadable) UpdateBytes(bytes []byte) {
	d.Bytes = bytes
}

func DownloadFileBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("bad status: %s", resp.Status))
	}
	var data bytes.Buffer
	_, err = io.Copy(&data, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func DownloadMultiple(values []*Downloadable) ([]*Downloadable, error) {
	done := make(chan *Downloadable, len(values))
	errs := make(chan error, len(values))
	throttle := time.Tick(time.Second / util.RequestPerSecond)
	for _, value := range values {
		go func(value *Downloadable) {
			<-throttle
			b, err := DownloadFileBytes(value.Url)
			if err != nil {
				errs <- err
				done <- nil
				return
			}
			value.UpdateBytes(b)
			done <- value
			errs <- nil
		}(value)
	}
	var errStr string
	for i := 0; i < len(values); i++ {
		if err := <-errs; err != nil {
			errStr = errStr + " " + err.Error()
		}
	}
	var err error
	if errStr != "" {
		err = errors.New(errStr)
	}
	return values, err
}

// GetTransactionReportMembers returns a slice of members that have transaction reports
// This is the list of Members for which to download PDF files
func GetTransactionReportMembers(downloads []*DisclosureDownload, dataFolder string) ([]*model.Member, error) {
	downloadMembers := make([]*model.Member, 0)
	var err error
	for _, disclosureDownload := range downloads {
		xmlPath := disclosureDownload.XmlPath
		var disclosure *model.FinancialDisclosure
		disclosure, err = model.CreateFinancialDisclosure(xmlPath)
		if err != nil {
			return downloadMembers, err
		}
		for _, member := range disclosure.Members {
			if member.ShouldDownload(dataFolder) {
				downloadMembers = append(downloadMembers, member)
			}
		}
	}
	return downloadMembers, err
}
