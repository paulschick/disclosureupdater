package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/s3client"
	"github.com/urfave/cli/v2"
)

func UploadPdfs(commonDirs *config.CommonDirs) CliFunc {
	var err error
	return func(cCtx *cli.Context) error {
		s3Profile := config.S3ProfileFromConfig(commonDirs, "default")
		var service *s3client.S3ServiceV2
		service, err = s3client.NewS3ServiceV2(s3Profile)
		if err != nil {
			fmt.Printf("Error creating S3ServiceV2 instance: %s\n", err)
			return err
		}
		fmt.Printf("Writing current bucket objects\n")
		err = service.WriteBucketObjects(commonDirs)
		if err != nil {
			fmt.Printf("Error writing bucket objects: %s\n", err)
			return err
		}
		fmt.Printf("Upload non-present PDF objects complete\n")
		return err
	}
}
