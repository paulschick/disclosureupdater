package model

import (
	"encoding/xml"
	"io"
	"os"
	"strconv"
)

const BasePdfUrl = "https://disclosures-clerk.house.gov/public_disc/"

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
	XMLName xml.Name `xml:"FinancialDisclosure"`
	Members []Member `xml:"Member"`
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

func (m Member) BuildPdfUrl() string {
	if m.FilingType == "P" {
		return BasePdfUrl + "ptr-pdfs/" + strconv.Itoa(m.Year) + "/" + strconv.Itoa(m.DocId) + ".pdf"
	} else {
		return BasePdfUrl + "financial-pdfs/" + strconv.Itoa(m.Year) + "/" + strconv.Itoa(m.DocId) + ".pdf"
	}
}

func (m Member) BuildPdfFileName() string {
	filingExtension := ".financial-pdfs"
	if m.FilingType == "P" {
		filingExtension = ".ptr-pdfs"
	}
	return strconv.Itoa(m.Year) + filingExtension + "." + m.StateDst +
		"." + m.Last + "." + m.First + "." + strconv.Itoa(m.DocId) + ".pdf"
}
