package cmds

import (
	"fmt"
	"github.com/otiai10/gosseract/v2"
	"github.com/paulschick/disclosureupdater/config"
	"github.com/paulschick/disclosureupdater/model"
	"github.com/urfave/cli/v2"
	"strings"
)

func OcrImages(commonDirs *config.CommonDirs) model.CliFunc {
	testImagePath := "/home/paul/software/APIs/disclosures-batch-updater/temp/test.png"

	return func(c *cli.Context) error {
		var err error
		var client *gosseract.Client
		client, err = initializeClient(testImagePath)
		if err != nil {
			return err
		}
		rows := extractRows(client)
		for _, row := range rows {
			tsvRow := row.ToTsvRow()
			fmt.Printf("LineNum\tWordNum\tWord\tConfidence\n")
			fmt.Printf("%s\n", tsvRow)
		}
		return nil
	}
}

type Row struct {
	LineNum    int
	WordNum    int
	Word       string
	Confidence float64
}

func NewRow(box gosseract.BoundingBox) *Row {
	return &Row{
		LineNum:    box.LineNum,
		WordNum:    box.WordNum,
		Word:       strings.ReplaceAll(strings.ReplaceAll(box.Word, "\n", ""), " ", ""),
		Confidence: box.Confidence,
	}
}

func (r *Row) ToTsvRow() string {
	return fmt.Sprintf("%d\t%d\t%s\t%f\n", r.LineNum, r.WordNum, r.Word, r.Confidence)
}

func initializeClient(imagePath string) (*gosseract.Client, error) {
	var err error
	client := gosseract.NewClient()
	err = client.SetImage(imagePath)
	if err != nil {
		fmt.Printf("Error setting image: %s\n", err.Error())
		return nil, err
	}
	err = client.SetTessdataPrefix("/usr/share/tesseract-ocr/5/tessdata_best-4.1.0")
	if err != nil {
		fmt.Printf("Error setting tessdata prefix: %s\n", err.Error())
		return nil, err
	}
	err = client.SetLanguage("eng")
	if err != nil {
		fmt.Printf("Error setting language: %s\n", err.Error())
		return nil, err
	}
	return client, nil
}

func extractRows(client *gosseract.Client) []*Row {
	out, err := client.GetBoundingBoxesVerbose()
	if err != nil {
		fmt.Printf("Error getting bounding boxes: %s\n", err.Error())
		return nil
	}
	rows := make([]*Row, 0)
	for _, box := range out {
		row := NewRow(box)
		rows = append(rows, row)
	}
	return rows
}
