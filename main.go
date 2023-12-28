package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/paulschick/disclosureupdater/s3client"
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

const (
	ZipUrlTemplate   = "https://disclosures-clerk.house.gov/public_disc/financial-pdfs/{YEAR}FD.zip"
	MinYear          = 2008
	RequestPerSecond = 100
)

func GenerateZipUrlForYear(year int) string {
	return strings.Replace(ZipUrlTemplate, "{YEAR}", strconv.Itoa(year), 1)
}

func CurrentYear() int {
	return time.Now().Year()
}

func GenerateAllZipUrls() []string {
	year := CurrentYear()
	numYears := year - MinYear + 1

	downloadUrls := make([]string, numYears)

	for i := MinYear; i <= year; i++ {
		downloadUrls[i-MinYear] = GenerateZipUrlForYear(i)
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
	urlFileName := path.Base(url)
	fileName := strings.Replace(urlFileName, ".zip", "", 1)
	zipPath := path.Join(baseFolder, fmt.Sprintf("%s.zip", fileName))
	xmlPath := path.Join(baseFolder, fmt.Sprintf("%s.xml", fileName))
	csvPath := path.Join(baseFolder, fmt.Sprintf("%s.csv", fileName))
	tempFilePath := path.Join(os.TempDir(), urlFileName)
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
	url   string
	bytes []byte
	fp    string
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

func CreatePdfDownloadDirectory(dataFolder string) error {
	return TryCreateDirectories(path.Join(dataFolder, model.BasePdfDir))
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

func TryCreateDirectories(fp string) (err error) {
	if _, err = os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(fp, os.ModePerm)
	}
	return err
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

func (d *Downloadable) UpdateBytes(bytes []byte) {
	d.bytes = bytes
}

func DownloadMultiple(values []*Downloadable) ([]*Downloadable, error) {
	done := make(chan *Downloadable, len(values))
	errs := make(chan error, len(values))
	throttle := time.Tick(time.Second / RequestPerSecond)
	for _, value := range values {
		go func(value *Downloadable) {
			<-throttle
			b, err := DownloadFileBytes(value.url)
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

func main() {
	fmt.Println("House of Representatives Data Updater")
	configuration := model.Configure()
	// TODO add to CLI program
	//
	//// Create directory if it does not exist, including subdirs
	//err := TryCreateDirectories(configuration.DataFolder)
	//if err != nil {
	//	panic(err)
	//}
	//
	//downloadUrls := GenerateAllZipUrls()
	//disclosureDownloads := make([]*DisclosureDownload, len(downloadUrls))
	//
	//fmt.Println(downloadUrls)
	//
	//for i := 0; i < len(downloadUrls); i++ {
	//	disclosureDownloads[i] = NewDisclosureDownload(downloadUrls[i], configuration.DataFolder)
	//	fmt.Println(disclosureDownloads[i].ToString())
	//}
	//
	//err = DownloadZipsIfNotPresent(disclosureDownloads)
	//if err != nil {
	//	panic(err)
	//}
	//
	//err = CreatePdfDownloadDirectory(configuration.DataFolder)
	//if err != nil {
	//	panic(err)
	//}
	//var downloadMembers []*model.Member
	//downloadMembers, err = GetTransactionReportMembers(disclosureDownloads, configuration.DataFolder)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Downloading %d PDFs\n", len(downloadMembers))
	//
	//downloadables := make([]*Downloadable, len(downloadMembers))
	//for i, member := range downloadMembers {
	//	fmt.Printf("Downloading %s\n", member.BuildPdfUrl())
	//	downloadables[i] = &Downloadable{
	//		url:   member.BuildPdfUrl(),
	//		bytes: nil,
	//		fp:    member.BuildPdfFilePath(configuration.DataFolder),
	//	}
	//}
	//downloadables, err = DownloadMultiple(downloadables)
	//if err != nil {
	//	log.Fatalf("Error downloading PDFs: %s", err)
	//}
	//for _, download := range downloadables {
	//	err = os.WriteFile(download.fp, download.bytes, 0644)
	//	if err != nil {
	//		log.Fatalf("Error writing PDF: %s", err)
	//	}
	//}
	// TODO end add to CLI program

	// TODO deprecate this section
	//err = S3(configuration)
	//if err != nil {
	//	log.Fatalf("Error uploading to S3: %s", err)
	//} else {
	//	fmt.Println("Successfully uploaded to S3")
	//}
	// TODO end deprecation section

	// TODO S3 service testing
	s3Service, err := s3client.NewS3Service(configuration)
	if err != nil {
		log.Fatalf("Error creating S3 service: %s", err)
	}
	err = s3Service.WriteBucketObjects()
	if err != nil {
		log.Fatalf("Error writing bucket objects: %s", err)
	} else {
		fmt.Println("Successfully wrote bucket objects")
	}
	// TODO end S3 service testing
}
