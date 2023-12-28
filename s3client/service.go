package s3client

import (
	"bufio"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/paulschick/disclosureupdater/model"
	"log"
	"os"
)

type S3Service struct {
	Client         *s3.Client
	C              *model.Config
	ObjectsPerPage int32
}

func NewS3Service(configuration *model.Config) (*S3Service, error) {
	var err error
	endpoint := aws.Endpoint{
		URL: configuration.S3Hostname,
	}
	endpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == configuration.S3Region {
			return endpoint, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	var cfg aws.Config
	cfg, err = config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedConfigProfile("default"),
		config.WithRegion(configuration.S3Region),
		config.WithEndpointResolverWithOptions(endpointResolver))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3Service{
		Client:         client,
		C:              configuration,
		ObjectsPerPage: 1000,
	}, err
}

func (s *S3Service) WriteBucketObjects() error {
	var maxKeys int32 = 1000
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.C.S3Bucket),
	}
	p := s3.NewListObjectsV2Paginator(s.Client, params, func(o *s3.ListObjectsV2PaginatorOptions) {
		o.Limit = maxKeys
	})
	pageNo := 0
	for p.HasMorePages() {
		log.Printf("Retrieving object from page %d\n", pageNo)
		page, err := p.NextPage(context.TODO())
		items := make([]string, 0)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			items = append(items, *obj.Key)
		}

		if pageNo == 0 {
			if _, b := os.Stat("s3_objects.txt"); !errors.Is(b, os.ErrNotExist) {
				// delete file
				err := os.Remove("s3_objects.txt")
				if err != nil {
					return err
				}
			}
			file, err := os.OpenFile("s3_objects.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			dataWriter := bufio.NewWriter(file)
			for _, item := range items {
				_, _ = dataWriter.WriteString(item + "\n")
			}
			_ = dataWriter.Flush()
			_ = file.Close()
		} else {
			// append to file
			file, err := os.OpenFile("s3_objects.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			dataWriter := bufio.NewWriter(file)
			for _, item := range items {
				_, _ = dataWriter.WriteString(item + "\n")
			}
			_ = dataWriter.Flush()
			_ = file.Close()
		}

		// first page should overwrite
		// following pages should append to existing file
		pageNo++
	}
	return nil

}

//func S3UploadFile(fp, bucket string, client *s3.Client) error {
//	f, err := os.Open(fp)
//	if err != nil {
//		log.Printf("failed to open file %q, %v", fp, err)
//		return errors.New(fmt.Sprintf("failed to open file %q, %v", fp, err))
//	}
//	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
//		Bucket: aws.String(bucket),
//		Key:    aws.String(path.Base(fp)),
//		Body:   f,
//	})
//
//	if err != nil {
//		log.Printf("failed to upload file, %v", err)
//		return errors.New(fmt.Sprintf("failed to upload file, %v", err))
//	}
//
//	err = f.Close()
//	if err != nil {
//		log.Printf("failed to close file %q, %v", fp, err)
//		return errors.New(fmt.Sprintf("failed to close file %q, %v", fp, err))
//	}
//	return nil
//}

//func Contains(key string, values []string) bool {
//	for _, value := range values {
//		if value == key {
//			return true
//		}
//	}
//	return false
//}
//
//func ContainsS3Object(key string, objects []types.Object) bool {
//	for _, object := range objects {
//		if *object.Key == key {
//			return true
//		}
//	}
//	return false
//}

// S3
// https://dev.to/aws-builders/get-objects-from-aws-s3-bucket-with-golang-2mne
// Uploads files to S3 bucket that are not present
//func S3(configuration *Config) error {
//	client, err := S3Client(configuration)
//	if err != nil {
//		return errors.New(fmt.Sprintf("failed to create s3 client, %v", err))
//	}
//	objects, err := ListS3Objects(configuration, client)
//	if err != nil {
//		return errors.New(fmt.Sprintf("failed to list objects, %v", err))
//	}
//	fmt.Printf("Found %d objects\n", len(objects))
//	var uploadedObjects []string
//	for _, object := range objects {
//		fmt.Printf("Found object %s\n", *object.Key)
//		keyCleaned := strings.Replace(*object.Key, "/", "", -1)
//		keyCleaned = strings.Replace(keyCleaned, " ", "", -1)
//		keyCleaned = strings.Replace(keyCleaned, "\n", "", -1)
//		keyCleaned = strings.Replace(keyCleaned, "\r", "", -1)
//		uploadedObjects = append(uploadedObjects, keyCleaned)
//		//uploadedObjects = append(uploadedObjects, *object.Key)
//	}
//
//	disclosureDir := path.Join(configuration.DataFolder, model.BasePdfDir)
//	files, err := os.ReadDir(disclosureDir)
//	if err != nil {
//		return errors.New(fmt.Sprintf("failed to read dir, %v", err))
//	}
//	//done := make(chan bool, len(files))
//	errs := make(chan error, len(files))
//	//throttle := time.Tick(time.Second / RequestPerSecond)
//	for _, file := range files {
//		fmt.Printf("Found file %s\n", file.Name())
//		if !ContainsS3Object(file.Name(), objects) && strings.Contains(file.Name(), ".pdf") {
//			fmt.Printf("Uploading file %s\n", file.Name())
//			fmt.Printf("Contains: %t\n", Contains(file.Name(), uploadedObjects))
//			fmt.Printf("Objects: %s\n", uploadedObjects[:10])
//			//	go func(file os.DirEntry) {
//			//		<-throttle
//			//		filePath := path.Join(disclosureDir, file.Name())
//			//		err = S3UploadFile(filePath, configuration.S3Bucket, client)
//			//		if err != nil {
//			//			errs <- err
//			//			done <- false
//			//		} else {
//			//			done <- true
//			//			errs <- nil
//			//		}
//			//	}(file)
//		}
//	}
//	var errStr string
//	for i := 0; i < len(files); i++ {
//		if err := <-errs; err != nil {
//			errStr = errStr + " " + err.Error()
//		}
//	}
//	if errStr != "" {
//		err = errors.New(errStr)
//	}
//	return err
//}
