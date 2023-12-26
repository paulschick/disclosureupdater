package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
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
	ZipUrlTemplate = "https://disclosures-clerk.house.gov/public_disc/financial-pdfs/{YEAR}FD.zip"
	MinYear        = 2008
)

func GenerateZipUrlForYear(year int) string {
	return strings.Replace(ZipUrlTemplate, "{YEAR}", strconv.Itoa(year), 1)
}

type Config struct {
	DbPath     string
	DataFolder string
}

func Configure() *Config {
	err := godotenv.Load("./.env")
	if err != nil {
		panic(err)
	}
	return &Config{
		DbPath: func() string {
			dbPath := os.Getenv("DB_PATH")
			if dbPath == "" {
				return "file::memory:?cache=shared"
			}
			return dbPath
		}(),
		DataFolder: func() string {
			dataFolder := os.Getenv("DATA_FOLDER")
			if dataFolder == "" {
				return "./data"
			}
			return dataFolder
		}(),
	}
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

func TryCreateDirectories(fp string) (err error) {
	if _, err = os.Stat(fp); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(fp, os.ModePerm)
	}
	return err
}

func main() {
	fmt.Println("House of Representatives Data Updater")
	config := Configure()
	fmt.Printf("DB Path: %s\n", config.DbPath)
	fmt.Printf("Data Folder: %s\n", config.DataFolder)

	// Create directory if it does not exist, including subdirs
	err := TryCreateDirectories(config.DataFolder)
	if err != nil {
		panic(err)
	}

	downloadUrls := GenerateAllZipUrls()
	disclosureDownloads := make([]*DisclosureDownload, len(downloadUrls))

	fmt.Println(downloadUrls)

	for i := 0; i < len(downloadUrls); i++ {
		disclosureDownloads[i] = NewDisclosureDownload(downloadUrls[i], config.DataFolder)
		fmt.Println(disclosureDownloads[i].ToString())
	}

	err = DownloadZipsIfNotPresent(disclosureDownloads)
	if err != nil {
		panic(err)
	}

	// Testing XML
	//xmlPath := disclosureDownloads[1].XmlPath
	//disclosure, err := model.CreateFinancialDisclosure(xmlPath)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("Members: %d\n", len(disclosure.Members))
	//firstMember := disclosure.Members[0]
	//fmt.Printf("First Member: %s %s\n", firstMember.First, firstMember.Last)
	//downloadUrl := firstMember.BuildPdfUrl()
	//fmt.Printf("Download URL: %s\n", downloadUrl)
	//fmt.Println("Year: " + strconv.Itoa(firstMember.Year) + " DocId: " + strconv.Itoa(firstMember.DocId))
	//fmt.Println(firstMember.DocId)
	//
	//// Test download pdf
	//memberPdfName := strconv.Itoa(firstMember.Year) + "." + firstMember.FilingType + "." + firstMember.StateDst +
	//	"." + strconv.Itoa(firstMember.DocId) + ".pdf"
	//fmt.Println(memberPdfName)
	//memberPdfName2 := firstMember.BuildPdfFileName()
	//fmt.Println(memberPdfName2)
	//
	//pdfPath := path.Join(config.DataFolder, "disclosures/"+memberPdfName2)
	//fmt.Println(pdfPath)
	//err = TryCreateDirectories(path.Dir(pdfPath))
	//if err != nil {
	//	panic(err)
	//}
	//// TODO - only download if transactions disclosure
	//// Also, figure out if I can do multiple downloads at once with goroutines
	//err = DownloadFile(downloadUrl, pdfPath)
	//if err != nil {
	//	panic(err)
	//}
}
