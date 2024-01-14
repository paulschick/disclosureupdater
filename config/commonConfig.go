package config

import (
	"github.com/paulschick/disclosureupdater/util"
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
	TestRotation      string
}

func NewCommonDirs(baseFolder string) *CommonDirs {
	dataFolder := path.Join(baseFolder, "data")
	return &CommonDirs{
		BaseFolder:        baseFolder,
		DataFolder:        dataFolder,
		S3Folder:          path.Join(dataFolder, "s3"),
		DisclosuresFolder: path.Join(dataFolder, "disclosures"),
		ImageFolder:       path.Join(dataFolder, "images"),
		OcrFolder:         path.Join(dataFolder, "ocr"),
		TestRotation:      path.Join(dataFolder, "test-rotation"),
	}
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
	err = util.TryCreateDirectories(c.TestRotation)
	if err != nil {
		return err
	}
	return nil
}

type S3Profile interface {
	GetBucket() string
	GetRegion() string
	GetHostname() string
	StaticAuthentication() bool
}

// S3DefaultProfile
// The profile used when credentials are initialized in the .aws directory
type S3DefaultProfile struct {
	S3Bucket   string
	S3Region   string
	S3Hostname string
}

func (s *S3DefaultProfile) GetBucket() string {
	return s.S3Bucket
}

func (s *S3DefaultProfile) GetRegion() string {
	return s.S3Region
}

func (s *S3DefaultProfile) GetHostname() string {
	return s.S3Hostname
}

func (s *S3DefaultProfile) StaticAuthentication() bool {
	return false
}

// S3StaticProfile
// The profile used when credentials are statically used from the config.yaml file
type S3StaticProfile struct {
	S3DefaultProfile
	S3ApiKey    string
	S3SecretKey string
}

func (s *S3StaticProfile) StaticAuthentication() bool {
	return true
}

func S3ProfileFromConfig(dirs *CommonDirs, profile string) S3Profile {
	err, v := ViperFromConfig(dirs)
	if err != nil {
		panic(err)
	}
	s3Bucket := v.GetString(profile + ".s3.s3bucket")
	s3Region := v.GetString(profile + ".s3.s3region")
	s3Hostname := v.GetString(profile + ".s3.s3hostname")
	s3Default := S3DefaultProfile{
		S3Bucket:   s3Bucket,
		S3Region:   s3Region,
		S3Hostname: s3Hostname,
	}
	s3ApiKey := v.GetString(profile + ".s3.s3ApiKey")
	s3SecretKey := v.GetString(profile + ".s3.s3SecretKey")
	if s3ApiKey != "" && s3SecretKey != "" {
		return &S3StaticProfile{
			S3DefaultProfile: s3Default,
			S3ApiKey:         s3ApiKey,
			S3SecretKey:      s3SecretKey,
		}
	} else {
		return &s3Default
	}
}

func S3ProfileFromCtx(c *cli.Context) S3Profile {
	s3Bucket := c.String("s3-bucket")
	s3Region := c.String("s3-region")
	s3Hostname := c.String("s3-hostname")
	s3Default := S3DefaultProfile{
		S3Bucket:   s3Bucket,
		S3Region:   s3Region,
		S3Hostname: s3Hostname,
	}

	// check if we have static credentials
	s3ApiKey := c.String("s3-api-key")
	s3SecretKey := c.String("s3-secret-key")
	if s3ApiKey != "" && s3SecretKey != "" {
		return &S3StaticProfile{
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

func BuildViper(profile string, dirs *CommonDirs, s3Profile S3Profile) *viper.Viper {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path.Join(dirs.BaseFolder, "config.yaml"))
	v.Set(profile+".dataFolder", dirs.DataFolder)
	v.Set(profile+".s3.s3Bucket", s3Profile.GetBucket())
	v.Set(profile+".s3.s3Region", s3Profile.GetRegion())
	v.Set(profile+".s3.s3Hostname", s3Profile.GetHostname())
	if s3Profile.StaticAuthentication() {
		v.Set(profile+".s3.s3ApiKey", s3Profile.(*S3StaticProfile).S3ApiKey)
		v.Set(profile+".s3.s3SecretKey", s3Profile.(*S3StaticProfile).S3SecretKey)
	}
	return v
}

func UpdateS3Config(s3Profile S3Profile, dirs *CommonDirs) error {
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
	if s3Profile.StaticAuthentication() {
		v.Set(profile+".s3.s3ApiKey", s3Profile.(*S3StaticProfile).S3ApiKey)
		v.Set(profile+".s3.s3SecretKey", s3Profile.(*S3StaticProfile).S3SecretKey)
	}
	err = v.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}
