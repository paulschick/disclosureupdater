package main

import (
	"fmt"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/s3client"
	"github.com/urfave/cli/v2"
)

func updateBucketItemIndex(commonDirs *config.CommonDirs) (*s3client.S3ServiceV2, error) {
	var err error
	s3Profile := config.S3ProfileFromConfig(commonDirs, "default")
	var service *s3client.S3ServiceV2
	service, err = s3client.NewS3ServiceV2(s3Profile)
	if err != nil {
		fmt.Printf("Error creating S3ServiceV2 instance: %s\n", err)
		return nil, err
	}
	fmt.Printf("Writing current bucket objects\n")
	err = service.WriteBucketObjects(commonDirs)
	if err != nil {
		fmt.Printf("Error writing bucket objects: %s\n", err)
		return nil, err
	}
	fmt.Println("Update bucket item index complete")
	return service, err
}

func UpdateBucketItemIndex(commonDirs *config.CommonDirs) CliFunc {
	return func(cCtx *cli.Context) error {
		_, err := updateBucketItemIndex(commonDirs)
		return err
	}
}

func UploadPdfs(commonDirs *config.CommonDirs) CliFunc {
	var err error
	return func(cCtx *cli.Context) error {
		shouldUpdateIndex := cCtx.Bool("update-index")
		var service *s3client.S3ServiceV2
		if shouldUpdateIndex {
			fmt.Printf("Updating bucket item index\n")
			service, err = updateBucketItemIndex(commonDirs)
			if err != nil {
				fmt.Printf("Error updating bucket item index: %s\n", err)
				return err
			}
		} else {
			fmt.Printf("Not updating bucket item index\n")
			s3Profile := config.S3ProfileFromConfig(commonDirs, "default")
			service, err = s3client.NewS3ServiceV2(s3Profile)
			if err != nil {
				fmt.Printf("Error creating S3ServiceV2 instance: %s\n", err)
				return err
			}
		}
		fmt.Printf("Operating on %s Bucket\n", service.S3Profile.GetBucket())

		err = service.UploadPdfsS3(commonDirs)

		// TODO - for reference, uploaded 5 files to get bucket index
		//fmt.Printf("DEMO - uploading 5 files to get a real index\n")

		//pdfDir := commonDirs.DisclosuresFolder
		//var files []os.DirEntry
		//files, err = os.ReadDir(pdfDir)
		//if err != nil {
		//	fmt.Printf("Error reading directory: %s\n", err)
		//	return err
		//}
		//for i, file := range files {
		//	if i > 4 {
		//		fmt.Printf("DEMO - only uploading 5 files - complete\n")
		//		break
		//	}
		//	fmt.Printf("DEMO - uploading file %d of %d\n", i+1, len(files))
		//	fmt.Printf("DEMO - uploading file %s\n", file.Name())
		//	//err = service.UploadFileToBucket(file)
		//	f, err := os.Open(fmt.Sprintf("%s/%s", pdfDir, file.Name()))
		//	if err != nil {
		//		fmt.Printf("Error opening file: %s\n", err)
		//		return err
		//	}
		//	err = service.UploadFile(f)
		//	if err != nil {
		//		fmt.Printf("Error uploading file: %s\n", err)
		//		return err
		//	}
		//	err = f.Close()
		//	if err != nil {
		//		fmt.Printf("Error closing file: %s\n", err)
		//		return err
		//	}
		//}

		return err
	}
}
