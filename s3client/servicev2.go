package s3client

import (
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
