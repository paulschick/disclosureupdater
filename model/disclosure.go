package model

import (
	"encoding/xml"
	"github.com/paulschick/disclosureupdater/common/constants"
	"io"
	"os"
	"strconv"
	"strings"
)

type Member struct {
	XMLName    xml.Name `xml:"Member"`
	Prefix     string   `xml:"Prefix"`
	Last       string   `xml:"Last"`
	First      string   `xml:"First"`
	Suffix     string   `xml:"Suffix"`
	FilingType string   `xml:"FilingType"`
	StateDst   string   `xml:"StateDst"`
	Year       int      `xml:"Year"`
	FilingDate string   `xml:"FilingDate"`
	DocId      int      `xml:"DocID"`
}

type FinancialDisclosure struct {
	XMLName xml.Name  `xml:"FinancialDisclosure"`
	Members []*Member `xml:"Member"`
}

func CreateFinancialDisclosure(xmlPath string) (*FinancialDisclosure, error) {
	xmlFile, err := os.Open(xmlPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		closeErr := xmlFile.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	bytes, _ := io.ReadAll(xmlFile)
	var disclosure FinancialDisclosure
	err = xml.Unmarshal(bytes, &disclosure)
	if err != nil {
		return nil, err
	}
	return &disclosure, nil
}

func (m *Member) BuildPdfUrl() string {
	if m.FilingType == "P" {
		return constants.BasePdfUrl + "ptr-pdfs/" + strconv.Itoa(m.Year) + "/" + strconv.Itoa(m.DocId) + ".pdf"
	} else {
		return constants.BasePdfUrl + "financial-pdfs/" + strconv.Itoa(m.Year) + "/" + strconv.Itoa(m.DocId) + ".pdf"
	}
}

func (m *Member) BuildPdfFileName() string {
	filingExtension := ".financial-pdfs"
	if m.FilingType == "P" {
		filingExtension = ".ptr-pdfs"
	}
	first := strings.Replace(m.First, " ", "_", -1)
	first = strings.Replace(first, ".", "", -1)
	last := strings.Replace(m.Last, " ", "_", -1)
	last = strings.Replace(last, ".", "", -1)
	return strconv.Itoa(m.Year) + filingExtension + "." + m.StateDst +
		"." + last + "." + first + "." + strconv.Itoa(m.DocId) + ".pdf"
}

func (m *Member) BuildPdfFilePath(dataFolder string) string {
	return dataFolder + "/" + constants.BasePdfDir + m.BuildPdfFileName()
}

func (m *Member) PdfFileExists(dataFolder string) bool {
	_, err := os.Stat(m.BuildPdfFilePath(dataFolder))
	return !os.IsNotExist(err)
}

func (m *Member) ShouldDownload(dataFolder string) bool {
	if m.FilingType == "P" && !m.PdfFileExists(dataFolder) {
		return true
	}
	return false
}
