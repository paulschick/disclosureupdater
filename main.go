package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joho/godotenv"
	"github.com/paulschick/disclosureupdater/model"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"slices"
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

type Config struct {
	DbPath     string
	DataFolder string
	S3Bucket   string
	S3Hostname string
	S3Region   string
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
		S3Bucket: func() string {
			s3Bucket := os.Getenv("S3_BUCKET")
			if s3Bucket == "" {
				return "house-ptr-202312-2"
			}
			return s3Bucket
		}(),
		S3Hostname: func() string {
			s3Hostname := os.Getenv("S3_HOSTNAME")
			if s3Hostname == "" {
				return "https://ewr1.vultrobjects.com"
			}
			return s3Hostname
		}(),
		S3Region: func() string {
			s3Region := os.Getenv("S3_REGION")
			if s3Region == "" {
				return "us-east-1"
			}
			return s3Region
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

func S3Client(configuration *Config) (*s3.Client, error) {
	endpoint := aws.Endpoint{
		URL: configuration.S3Hostname,
	}
	endpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == configuration.S3Region {
			return endpoint, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile("default"),
		config.WithRegion(configuration.S3Region),
		config.WithEndpointResolverWithOptions(endpointResolver))
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}

func testUpload(configuration *Config, client *s3.Client) {
	disclosureDir := path.Join(configuration.DataFolder, model.BasePdfDir)
	files, err := os.ReadDir(disclosureDir)
	if err != nil {
		panic(err)
	}

	//test upload with 5 files
	//ONLY DO ONCE PRIOR TO READING FROM BUCKET
	for i := 0; i < 5; i++ {
		file := files[i]
		filePath := path.Join(disclosureDir, file.Name())
		fmt.Printf("Uploading %s\n", filePath)
		f, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("failed to open file %q, %v", filePath, err)
		}

		_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(configuration.S3Bucket),
			Key:    aws.String(file.Name()),
			Body:   f,
		})
		if err != nil {
			log.Fatalf("failed to upload file, %v", err)
		}

		err = f.Close()
		if err != nil {
			log.Fatalf("failed to close file %q, %v", filePath, err)
		}
	}
}

func ListS3Objects(configuration *Config, client *s3.Client) ([]types.Object, error) {
	output, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(configuration.S3Bucket),
	})
	var contents []types.Object
	if err != nil {
		return nil, err
	} else {
		contents = output.Contents
	}
	return contents, err
}

// S3
// https://dev.to/aws-builders/get-objects-from-aws-s3-bucket-with-golang-2mne
// Uploads files to S3 bucket that are not present
func S3(configuration *Config) {
	client, err := S3Client(configuration)
	if err != nil {
		log.Fatalf("failed to create s3 client, %v", err)
	}
	objects, err := ListS3Objects(configuration, client)
	if err != nil {
		log.Fatalf("failed to list objects, %v", err)
	}
	fmt.Printf("Found %d objects\n", len(objects))
	var uploadedObjects []string
	for _, object := range objects {
		fmt.Printf("Found object %s\n", *object.Key)
		uploadedObjects = append(uploadedObjects, *object.Key)
	}

	disclosureDir := path.Join(configuration.DataFolder, model.BasePdfDir)
	files, err := os.ReadDir(disclosureDir)
	if err != nil {
		panic(err)
	}
	uploadedCount := 0
	for _, file := range files {
		if !slices.Contains(uploadedObjects, file.Name()) {
			if uploadedCount < 5 {
				fmt.Printf("Uploading %s\n", file.Name())
				filePath := path.Join(disclosureDir, file.Name())
				fmt.Printf("Uploading %s\n", filePath)
				f, err := os.Open(filePath)
				if err != nil {
					log.Fatalf("failed to open file %q, %v", filePath, err)
				}

				_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
					Bucket: aws.String(configuration.S3Bucket),
					Key:    aws.String(file.Name()),
					Body:   f,
				})
				if err != nil {
					log.Fatalf("failed to upload file, %v", err)
				}

				err = f.Close()
				if err != nil {
					log.Fatalf("failed to close file %q, %v", filePath, err)
				}

				uploadedCount++
			} else {
				fmt.Printf("Skipping %s\n", file.Name())
				break
			}
		}
	}

}

func main() {
	fmt.Println("House of Representatives Data Updater")
	configuration := Configure()
	fmt.Printf("DB Path: %s\n", configuration.DbPath)
	fmt.Printf("Data Folder: %s\n", configuration.DataFolder)

	// Create directory if it does not exist, including subdirs
	err := TryCreateDirectories(configuration.DataFolder)
	if err != nil {
		panic(err)
	}

	downloadUrls := GenerateAllZipUrls()
	disclosureDownloads := make([]*DisclosureDownload, len(downloadUrls))

	fmt.Println(downloadUrls)

	for i := 0; i < len(downloadUrls); i++ {
		disclosureDownloads[i] = NewDisclosureDownload(downloadUrls[i], configuration.DataFolder)
		fmt.Println(disclosureDownloads[i].ToString())
	}

	err = DownloadZipsIfNotPresent(disclosureDownloads)
	if err != nil {
		panic(err)
	}

	err = CreatePdfDownloadDirectory(configuration.DataFolder)
	if err != nil {
		panic(err)
	}
	var downloadMembers []*model.Member
	downloadMembers, err = GetTransactionReportMembers(disclosureDownloads, configuration.DataFolder)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Downloading %d PDFs\n", len(downloadMembers))

	downloadables := make([]*Downloadable, len(downloadMembers))
	for i, member := range downloadMembers {
		fmt.Printf("Downloading %s\n", member.BuildPdfUrl())
		downloadables[i] = &Downloadable{
			url:   member.BuildPdfUrl(),
			bytes: nil,
			fp:    member.BuildPdfFilePath(configuration.DataFolder),
		}
	}
	downloadables, err = DownloadMultiple(downloadables)
	if err != nil {
		log.Fatalf("Error downloading PDFs: %s", err)
	}
	for _, download := range downloadables {
		err = os.WriteFile(download.fp, download.bytes, 0644)
		if err != nil {
			log.Fatalf("Error writing PDF: %s", err)
		}
	}

	S3(configuration)
}
