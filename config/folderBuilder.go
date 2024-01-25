package config

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
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

type FolderBuilder interface {
	BuildCommonDirs() *CommonDirs
}

type CtxFolderBuilder struct {
	c              *cli.Context
	imagesKey      string
	disclosuresKey string
	ocrKey         string
	csvKey         string
}

func NewCtxFolderBuilder(c *cli.Context) *CtxFolderBuilder {
	return &CtxFolderBuilder{
		c:              c,
		imagesKey:      "images",
		disclosuresKey: "disclosures",
		ocrKey:         "ocr",
		csvKey:         "csv",
	}
}

func (f *CtxFolderBuilder) constructPath(folderName string) string {
	return path.Join(GetDataFolder(), folderName)
}

func (f *CtxFolderBuilder) getFolderFromCtx(flagName, defaultFolderName string) string {
	folder := f.c.String(flagName)
	if folder == "" {
		folder = defaultFolderName
	}
	return f.constructPath(folder)
}

func (f *CtxFolderBuilder) getImagesFolderCtx() string {
	return f.getFolderFromCtx(f.imagesKey, DefaultImageFolder)
}

func (f *CtxFolderBuilder) getDisclosuresFolderCtx() string {
	return f.getFolderFromCtx(f.disclosuresKey, DefaultDisclosuresFolder)
}

func (f *CtxFolderBuilder) getOcrFolderCtx() string {
	return f.getFolderFromCtx(f.ocrKey, DefaultOcrFolder)
}

func (f *CtxFolderBuilder) getCsvFolderCtx() string {
	return f.getFolderFromCtx(f.csvKey, DefaultCsvFolder)
}

func (f *CtxFolderBuilder) getS3Folder() string {
	return f.constructPath(DefaultS3Folder)
}

func (f *CtxFolderBuilder) BuildCommonDirs() *CommonDirs {
	return &CommonDirs{
		BaseFolder:        GetBaseFolder(),
		DataFolder:        GetDataFolder(),
		S3Folder:          f.getS3Folder(),
		DisclosuresFolder: f.getDisclosuresFolderCtx(),
		ImageFolder:       f.getImagesFolderCtx(),
		OcrFolder:         f.getOcrFolderCtx(),
		CsvFolder:         f.getCsvFolderCtx(),
	}
}

type ViperFolderBuilder struct {
	v       *viper.Viper
	profile string
}

func NewViperFolderBuilder(v *viper.Viper, profile string) *ViperFolderBuilder {
	return &ViperFolderBuilder{
		v:       v,
		profile: profile,
	}
}

func (f *ViperFolderBuilder) getFolderFromViper(key, defaultFolderName string) string {
	folder := f.v.GetString(f.profile + "." + key)
	if folder == "" {
		folder = defaultFolderName
	}
	return path.Join(GetDataFolder(), folder)
}

func (f *ViperFolderBuilder) getImagesFolderViper() string {
	return f.getFolderFromViper("data.imagesFolder", DefaultImageFolder)
}

func (f *ViperFolderBuilder) getDisclosuresFolderViper() string {
	return f.getFolderFromViper("data.disclosuresFolder", DefaultDisclosuresFolder)
}

func (f *ViperFolderBuilder) getOcrFolderViper() string {
	return f.getFolderFromViper("data.ocrFolder", DefaultOcrFolder)
}

func (f *ViperFolderBuilder) getCsvFolderViper() string {
	return f.getFolderFromViper("data.csvFolder", DefaultCsvFolder)
}

func (f *ViperFolderBuilder) getS3Folder() string {
	return path.Join(GetDataFolder(), DefaultS3Folder)
}

func (f *ViperFolderBuilder) BuildCommonDirs() *CommonDirs {
	return &CommonDirs{
		BaseFolder:        GetBaseFolder(),
		DataFolder:        GetDataFolder(),
		S3Folder:          f.getS3Folder(),
		DisclosuresFolder: f.getDisclosuresFolderViper(),
		ImageFolder:       f.getImagesFolderViper(),
		OcrFolder:         f.getOcrFolderViper(),
		CsvFolder:         f.getCsvFolderViper(),
	}
}
