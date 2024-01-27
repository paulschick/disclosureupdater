package constants

const (
	ZipUrlTemplate           = "https://disclosures-clerk.house.gov/public_disc/financial-pdfs/{YEAR}FD.zip"
	MinYear                  = 2008
	RequestPerSecond         = 100
	MaxConversions           = 25
	BasePdfUrl               = "https://disclosures-clerk.house.gov/public_disc/"
	BasePdfDir               = "disclosures/"
	DefaultImageFolder       = "images"
	DefaultDisclosuresFolder = "disclosures"
	DefaultOcrFolder         = "ocr"
	DefaultCsvFolder         = "csv"
	DefaultS3Folder          = "s3"
	DefaultProfile           = "default"
	MaxJobs                  = 25
	CpuUtilization           = 0.7
	BatchSize                = 100
)
