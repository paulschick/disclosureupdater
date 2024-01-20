package model

import (
	"github.com/otiai10/gosseract/v2"
	"strings"
)

type OcrResult struct {
	LineNum    int     `csv:"lineNum"`
	WordNum    int     `csv:"wordNum"`
	Word       string  `csv:"word"`
	Confidence float64 `csv:"confidence"`
}

func NewOcrResult(box gosseract.BoundingBox) *OcrResult {
	return &OcrResult{
		LineNum:    box.LineNum,
		WordNum:    box.WordNum,
		Word:       strings.ReplaceAll(strings.ReplaceAll(box.Word, "\n", ""), " ", ""),
		Confidence: box.Confidence,
	}
}
