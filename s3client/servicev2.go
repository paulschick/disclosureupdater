package s3client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
