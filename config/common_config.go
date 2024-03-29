package config

import (
	"github.com/paulschick/disclosureupdater/common/constants"
	"github.com/paulschick/disclosureupdater/common/methods"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
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

func (c *CommonDirs) ToString() string {
	return "CommonDirs{\n" +
		"\tBaseFolder: " + c.BaseFolder + ",\n" +
		"\tDataFolder: " + c.DataFolder + ",\n" +
		"\tS3Folder: " + c.S3Folder + ",\n" +
		"\tDisclosuresFolder: " + c.DisclosuresFolder + ",\n" +
		"\tImageFolder: " + c.ImageFolder + ",\n" +
		"\tOcrFolder: " + c.OcrFolder + ",\n" +
		"\tCsvFolder: " + c.CsvFolder + ",\n" +
		"}"
}

func NewCommonDirsFromCtx(c *cli.Context) *CommonDirs {
	dataFolder := GetDataFolder()
	imagesFolder := c.String("images")
	if imagesFolder == "" {
		imagesFolder = path.Join(dataFolder, constants.DefaultImageFolder)
	}
	disclosuresFolder := c.String("disclosures")
	if disclosuresFolder == "" {
		disclosuresFolder = path.Join(dataFolder, constants.DefaultDisclosuresFolder)
	}
	ocrFolder := c.String("ocr")
	if ocrFolder == "" {
		ocrFolder = path.Join(dataFolder, constants.DefaultOcrFolder)
	}
	csvFolder := c.String("csv")
	if csvFolder == "" {
		csvFolder = path.Join(dataFolder, constants.DefaultCsvFolder)
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

func (c *CommonDirs) CreateDirectories() error {
	dirs := []string{
		c.S3Folder,
		c.DisclosuresFolder,
		c.ImageFolder,
		c.OcrFolder,
		c.CsvFolder,
	}
	for _, dir := range dirs {
		if err := methods.TryCreateDirectories(dir); err != nil {
			return err
		}
	}
	return nil
}

func S3ProfileFromConfig(profile string) model.S3Profile {
	v, err := InitializeViper()
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

func InitializeViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(GetBaseFolder(), GetConfigFileName()))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
