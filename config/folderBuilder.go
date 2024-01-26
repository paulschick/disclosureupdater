package config

import (
	"os"
	"path"
)

const (
	DefaultImageFolder       = "images"
	DefaultDisclosuresFolder = "disclosures"
	DefaultOcrFolder         = "ocr"
	DefaultCsvFolder         = "csv"
	DefaultS3Folder          = "s3"
)

func GetBaseFolder() string {
	return path.Join(os.Getenv("HOME"), ".disclosurecli")
}

func GetDataFolder() string {
	return path.Join(GetBaseFolder(), "data")
}

// GetConfigFileName Currently just returns default
// allow this to be configurable
func GetConfigFileName() string {
	return "config.yaml"
}

// GetConfigProfile Currently just returns default
func GetConfigProfile() string {
	return "default"
}
