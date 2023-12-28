package util

import "time"

const (
	ZipUrlTemplate   = "https://disclosures-clerk.house.gov/public_disc/financial-pdfs/{YEAR}FD.zip"
	MinYear          = 2008
	RequestPerSecond = 100
)

func CurrentYear() int {
	return time.Now().Year()
}
