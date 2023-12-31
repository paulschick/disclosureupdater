package config

import (
	"github.com/paulschick/disclosureupdater/util"
	"github.com/spf13/viper"
	"path"
)

type CommonDirs struct {
	BaseFolder        string
	DataFolder        string
	S3Folder          string
	DisclosuresFolder string
}

func NewCommonDirs(baseFolder string) *CommonDirs {
	dataFolder := path.Join(baseFolder, "data")
	return &CommonDirs{
		BaseFolder:        baseFolder,
		DataFolder:        dataFolder,
		S3Folder:          path.Join(dataFolder, "s3"),
		DisclosuresFolder: path.Join(dataFolder, "disclosures"),
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
