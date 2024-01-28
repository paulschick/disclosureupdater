package config

import (
	"github.com/magiconair/properties/assert"
	"github.com/paulschick/disclosureupdater/model"
	"path"
	"testing"
)

func TestNewCommonDirs(t *testing.T) {
	baseFolder := "/tmp"
	expectedDataFolder := path.Join(baseFolder, "data")
	expectedS3Folder := path.Join(expectedDataFolder, "s3")
	expectedDisclosuresFolder := path.Join(expectedDataFolder, "disclosures")

	commonDirs := NewCommonDirs(baseFolder)
	assert.Equal(t, baseFolder, commonDirs.BaseFolder)
	assert.Equal(t, expectedDataFolder, commonDirs.DataFolder)
	assert.Equal(t, expectedS3Folder, commonDirs.S3Folder)
	assert.Equal(t, expectedDisclosuresFolder, commonDirs.DisclosuresFolder)
}

func TestCommonDirs_CreateDirectories(t *testing.T) {
	baseFolder := "/tmp"
	commonDirs := NewCommonDirs(baseFolder)
	err := commonDirs.CreateDirectories()
	assert.Equal(t, nil, err)
}

func TestS3Profile(t *testing.T) {
	s3Profile := &model.S3DefaultProfile{
		S3Bucket:   "bucket",
		S3Region:   "region",
		S3Hostname: "hostname",
	}
	assert.Equal(t, "bucket", s3Profile.GetBucket())
	assert.Equal(t, "region", s3Profile.GetRegion())
	assert.Equal(t, "hostname", s3Profile.GetHostname())
	assert.Equal(t, false, s3Profile.StaticAuthentication())
}

func TestS3ProfileStaticAuthentication(t *testing.T) {
	S3Profile := &model.S3StaticProfile{
		S3DefaultProfile: model.S3DefaultProfile{
			S3Bucket:   "bucket",
			S3Region:   "region",
			S3Hostname: "hostname",
		},
		S3ApiKey:    "apiKey",
		S3SecretKey: "secretKey",
	}
	assert.Equal(t, "bucket", S3Profile.GetBucket())
	assert.Equal(t, "region", S3Profile.GetRegion())
	assert.Equal(t, "hostname", S3Profile.GetHostname())
	assert.Equal(t, true, S3Profile.StaticAuthentication())
	assert.Equal(t, "apiKey", S3Profile.S3ApiKey)
	assert.Equal(t, "secretKey", S3Profile.S3SecretKey)
}

func TestBuildViper(t *testing.T) {
	commonDirs := NewCommonDirs("/tmp")
	s3Profile := &model.S3DefaultProfile{
		S3Bucket:   "bucket",
		S3Region:   "region",
		S3Hostname: "hostname",
	}
	v := BuildViper("default", commonDirs, s3Profile)
	assert.Equal(t, "/tmp/config.yaml", v.ConfigFileUsed())
	assert.Equal(t, "/tmp/data", v.GetString("default.dataFolder"))
}
