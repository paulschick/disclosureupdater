package model

import (
	"github.com/joho/godotenv"
	"os"
)

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
