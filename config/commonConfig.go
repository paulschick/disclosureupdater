package config

import (
	"github.com/paulschick/disclosureupdater/model"
	"github.com/paulschick/disclosureupdater/util"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"os"
	"path"
)

type CommonDirs struct {
	BaseFolder        string
	DataFolder        string
	S3Folder          string
	DisclosuresFolder string
	ImageFolder       string
	OcrFolder         string
	CsvFolder         string
}

const (
	DefaultImageFolder       = "images"
	DefaultDisclosuresFolder = "disclosures"
	DefaultOcrFolder         = "ocr"
	DefaultCsvFolder         = "csv"
)

func NewCommonDirsFromCtx(c *cli.Context) *CommonDirs {
	dataFolder := GetDataFolder()
	imagesFolder := c.String("images")
	if imagesFolder == "" {
		imagesFolder = path.Join(dataFolder, DefaultImageFolder)
	}
	disclosuresFolder := c.String("disclosures")
	if disclosuresFolder == "" {
		disclosuresFolder = path.Join(dataFolder, DefaultDisclosuresFolder)
	}
	ocrFolder := c.String("ocr")
	if ocrFolder == "" {
		ocrFolder = path.Join(dataFolder, DefaultOcrFolder)
	}
	csvFolder := c.String("csv")
	if csvFolder == "" {
		csvFolder = path.Join(dataFolder, DefaultCsvFolder)
	}
	return &CommonDirs{
		BaseFolder:        GetBaseFolder(),
		DataFolder:        dataFolder,
		S3Folder:          path.Join(dataFolder, "s3"),
		DisclosuresFolder: disclosuresFolder,
		ImageFolder:       imagesFolder,
		OcrFolder:         ocrFolder,
		CsvFolder:         csvFolder,
	}
}

func GetBaseFolder() string {
	return path.Join(os.Getenv("HOME"), ".disclosurecli")
}

func GetDataFolder() string {
	return path.Join(GetBaseFolder(), "data")
}

// NewCommonDirs
// TODO - allow these to be changed
func NewCommonDirs(baseFolder string) *CommonDirs {
	dataFolder := GetDataFolder()
	return &CommonDirs{
		BaseFolder:        baseFolder,
		DataFolder:        dataFolder,
		S3Folder:          path.Join(dataFolder, "s3"),
		DisclosuresFolder: path.Join(dataFolder, "disclosures"),
		ImageFolder:       path.Join(dataFolder, "images"),
		OcrFolder:         path.Join(dataFolder, "ocr"),
		CsvFolder:         path.Join(dataFolder, "csv"),
	}
}

func NewCommonDirsFromConfig() (*CommonDirs, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(GetBaseFolder(), "config.yaml"))
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	profile := "default"
	dataFolder := GetDataFolder()
	s3Folder := path.Join(dataFolder, "s3")
	disclosuresFolder := v.GetString(profile + ".data.disclosuresFolder")
	if disclosuresFolder == "" {
		disclosuresFolder = path.Join(dataFolder, "disclosures")
	}
	imageFolder := v.GetString(profile + ".data.imagesFolder")
	if imageFolder == "" {
		imageFolder = path.Join(dataFolder, "images")
	}
	ocrFolder := v.GetString(profile + ".data.ocrFolder")
	if ocrFolder == "" {
		ocrFolder = path.Join(dataFolder, "ocr")
	}
	csvFolder := v.GetString(profile + ".data.csvFolder")
	if csvFolder == "" {
		csvFolder = path.Join(dataFolder, "csv")
	}
	return &CommonDirs{
		BaseFolder:        GetBaseFolder(),
		DataFolder:        dataFolder,
		S3Folder:          s3Folder,
		DisclosuresFolder: disclosuresFolder,
		ImageFolder:       imageFolder,
		OcrFolder:         ocrFolder,
		CsvFolder:         csvFolder,
	}, nil
}

func (c *CommonDirs) CreateDirectories() error {
	err := util.TryCreateDirectories(c.S3Folder)
	if err != nil {
		return err
	}
	err = util.TryCreateDirectories(c.DisclosuresFolder)
	if err != nil {
		return err
	}
	err = util.TryCreateDirectories(c.ImageFolder)
	if err != nil {
		return err
	}
	err = util.TryCreateDirectories(c.OcrFolder)
	if err != nil {
		return err
	}
	err = util.TryCreateDirectories(c.CsvFolder)
	if err != nil {
		return err
	}
	return nil
}

func S3ProfileFromConfig(dirs *CommonDirs, profile string) model.S3Profile {
	err, v := ViperFromConfig(dirs)
	if err != nil {
		panic(err)
	}
	s3Bucket := v.GetString(profile + ".s3.s3bucket")
	s3Region := v.GetString(profile + ".s3.s3region")
	s3Hostname := v.GetString(profile + ".s3.s3hostname")
	s3Default := model.S3DefaultProfile{
		S3Bucket:   s3Bucket,
		S3Region:   s3Region,
		S3Hostname: s3Hostname,
	}
	s3ApiKey := v.GetString(profile + ".s3.s3ApiKey")
	s3SecretKey := v.GetString(profile + ".s3.s3SecretKey")
	if s3ApiKey != "" && s3SecretKey != "" {
		return &model.S3StaticProfile{
			S3DefaultProfile: s3Default,
			S3ApiKey:         s3ApiKey,
			S3SecretKey:      s3SecretKey,
		}
	} else {
		return &s3Default
	}
}

func S3ProfileFromCtx(c *cli.Context) model.S3Profile {
	s3Bucket := c.String("s3-bucket")
	s3Region := c.String("s3-region")
	s3Hostname := c.String("s3-hostname")
	s3Default := model.S3DefaultProfile{
		S3Bucket:   s3Bucket,
		S3Region:   s3Region,
		S3Hostname: s3Hostname,
	}

	// check if we have static credentials
	s3ApiKey := c.String("s3-api-key")
	s3SecretKey := c.String("s3-secret-key")
	if s3ApiKey != "" && s3SecretKey != "" {
		return &model.S3StaticProfile{
			S3DefaultProfile: s3Default,
			S3ApiKey:         s3ApiKey,
			S3SecretKey:      s3SecretKey,
		}
	} else {
		return &s3Default
	}
}

func ViperFromConfig(dirs *CommonDirs) (error, *viper.Viper) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(dirs.BaseFolder, "config.yaml"))
	err := v.ReadInConfig()
	if err != nil {
		return err, nil
	}
	return nil, v
}

func BuildViper(profile string, dirs *CommonDirs, s3Profile model.S3Profile) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(dirs.BaseFolder, "config.yaml"))
	v.Set(profile+".dataFolder", dirs.DataFolder)
	v.Set(profile+".s3.s3Bucket", s3Profile.GetBucket())
	v.Set(profile+".s3.s3Region", s3Profile.GetRegion())
	v.Set(profile+".s3.s3Hostname", s3Profile.GetHostname())
	if s3Profile.StaticAuthentication() {
		v.Set(profile+".s3.s3ApiKey", s3Profile.(*model.S3StaticProfile).S3ApiKey)
		v.Set(profile+".s3.s3SecretKey", s3Profile.(*model.S3StaticProfile).S3SecretKey)
	}
	return v
}

func UpdateS3Config(s3Profile model.S3Profile, dirs *CommonDirs) error {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(dirs.BaseFolder, "config.yaml"))
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	profile := "default"
	v.Set(profile+".s3.s3Bucket", s3Profile.GetBucket())
	v.Set(profile+".s3.s3Region", s3Profile.GetRegion())
	v.Set(profile+".s3.s3Hostname", s3Profile.GetHostname())
	v.Set(profile+".data.disclosuresFolder", dirs.DisclosuresFolder)
	v.Set(profile+".data.imagesFolder", dirs.ImageFolder)
	v.Set(profile+".data.ocrFolder", dirs.OcrFolder)
	v.Set(profile+".data.csvFolder", dirs.CsvFolder)
	if s3Profile.StaticAuthentication() {
		v.Set(profile+".s3.s3ApiKey", s3Profile.(*model.S3StaticProfile).S3ApiKey)
		v.Set(profile+".s3.s3SecretKey", s3Profile.(*model.S3StaticProfile).S3SecretKey)
	}
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
