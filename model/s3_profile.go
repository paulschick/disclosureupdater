package model

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
