package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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

func (d *DisclosureDownload) Download() error {
	response, err := http.Get(d.Url)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return err
	}
	file, err := os.Create(d.ZipPath)
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
	for _, disclosureDownload := range downloads {
		if !disclosureDownload.XmlIsPresent() {
			if !disclosureDownload.ZipIsPresent() {
				err = disclosureDownload.Download()
				if err != nil {
					return err
				}
				err = disclosureDownload.Extract()
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("Skipping download of %s\n", disclosureDownload.ZipPath)
				err = disclosureDownload.Extract()
				if err != nil {
					return err
				}
			}
		} else {
			fmt.Printf("Skipping download of %s\n", disclosureDownload.XmlPath)
		}
	}
	return nil
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
}
