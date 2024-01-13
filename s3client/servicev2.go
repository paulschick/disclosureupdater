package s3client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	conf "github.com/paulschick/disclosureupdater/config"
	"os"
	"path/filepath"
)

type S3ServiceV2 struct {
	Client    *s3.Client
	S3Profile conf.S3Profile
}

func NewS3ServiceV2(s3Profile conf.S3Profile) (*S3ServiceV2, error) {
	endpoint := aws.Endpoint{
		URL: s3Profile.GetHostname(),
	}
	endpointResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == s3Profile.GetRegion() {
			return endpoint, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	var cfg aws.Config
	var err error
	if s3Profile.StaticAuthentication() {
		apiKey := s3Profile.(*conf.S3StaticProfile).S3ApiKey
		apiSecret := s3Profile.(*conf.S3StaticProfile).S3SecretKey
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithRegion(s3Profile.GetRegion()),
			config.WithEndpointResolverWithOptions(endpointResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(apiKey, apiSecret, "")))
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithSharedConfigProfile("default"),
			config.WithRegion(s3Profile.GetRegion()),
			config.WithEndpointResolverWithOptions(endpointResolver))
		if err != nil {
			return nil, err
		}
	}
	client := s3.NewFromConfig(cfg)
	return &S3ServiceV2{
		Client:    client,
		S3Profile: s3Profile,
	}, err
}

func (s *S3ServiceV2) CreateNewBucket() error {
	exists, err := s.BucketExists()
	if err != nil {
		fmt.Printf("Error checking if bucket exists: %s\n", err)
		return err
	}
	if exists {
		return nil
	}
	_, err = s.Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(s.S3Profile.GetBucket()),
	})
	if err != nil {
		fmt.Printf("Error creating bucket: %s\n", err)
		return err
	}
	return nil
}

func (s *S3ServiceV2) BucketExists() (bool, error) {
	_, err := s.Client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(s.S3Profile.GetBucket()),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				exists = false
				err = nil
			default:
				return false, err
			}
		}
	} else {
		fmt.Printf("Bucket %s exists\n", s.S3Profile.GetBucket())
	}
	return exists, err
}

func (s *S3ServiceV2) WriteBucketObjects(commonDirs *conf.CommonDirs) error {
	var maxKeys int32 = 1000
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.S3Profile.GetBucket()),
	}
	p := s3.NewListObjectsV2Paginator(s.Client, params, func(o *s3.ListObjectsV2PaginatorOptions) {
		o.Limit = maxKeys
	})
	pageNo := 0
	for p.HasMorePages() {
		page, err := p.NextPage(context.TODO())
		items := make([]string, 0)
		if err != nil {
			return err
		}
		for _, obj := range page.Contents {
			items = append(items, *obj.Key)
		}
		fp := filepath.Join(commonDirs.S3Folder, "s3_objects.txt")
		var file *os.File
		if pageNo == 0 {
			if _, b := os.Stat(fp); !errors.Is(b, os.ErrNotExist) {
				// delete file
				err = os.Remove(fp)
				if err != nil {
					return err
				}
			}
		}
		file, err = os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		dataWriter := bufio.NewWriter(file)
		for _, item := range items {
			_, _ = dataWriter.WriteString(item + "\n")
		}
		_ = dataWriter.Flush()
		_ = file.Close()

		pageNo++
	}
	return nil
}

func (s *S3ServiceV2) UploadPdfsS3(commonDirs *conf.CommonDirs) error {
	var err error
	indexFp := filepath.Join(commonDirs.S3Folder, "s3_objects.txt")
	if _, b := os.Stat(indexFp); errors.Is(b, os.ErrNotExist) {
		fmt.Printf("No index file found at %s\n", indexFp)
		// there should be an empty file at least
		return nil
	}
	var file *os.File
	file, err = os.Open(indexFp)
	if err != nil {
		fmt.Printf("Error opening file %s: %s\n", indexFp, err)
		return err
	}
	defer func() {
		// not handling close error for now
		_ = file.Close()
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// This is a file name in S3
		line := scanner.Text()
		fmt.Printf("S3 Item:\t%s\n", line)
	}

	return err
}

func (s *S3ServiceV2) UploadFile(file *os.File) error {
	var err error
	fileName := file.Name()
	_, err = s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.S3Profile.GetBucket()),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		fmt.Printf("Failed to upload file %s: %s\n", fileName, err)
		return err
	}
	return err
}
