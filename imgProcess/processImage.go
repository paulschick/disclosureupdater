package imgProcess

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

// ProcessImage
// Resource: https://medium.com/@damithadayananda/image-processing-with-golang-8f20d2d243a2
type ProcessImage struct {
	SourcePath string
	OutputPath string
}

func NewProcessImage(sourcePath, outputPath string) *ProcessImage {
	return &ProcessImage{
		SourcePath: sourcePath,
		OutputPath: outputPath,
	}
}

func (p *ProcessImage) OpenImage() (image.Image, error) {
	f, err := os.Open(p.SourcePath)
	if err != nil {
		return nil, err
	}
	fi, _ := f.Stat()
	fmt.Printf("Opened File: %s\n", fi.Name())

	img, format, err := image.Decode(f)
	if err != nil {
		fmt.Println("Decoding Error: ", err.Error())
		return nil, err
	}
	if format != "png" {
		fmt.Printf("Image format: %s\n", format)
		return nil, errors.New("image format not supported")
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (p *ProcessImage) ImageToTensor(img image.Image) [][]color.Color {
	size := img.Bounds().Size()
	var pixels [][]color.Color
	for i := 0; i < size.X; i++ {
		var y []color.Color
		for j := 0; j < size.Y; j++ {
			y = append(y, img.At(i, j))
		}
		pixels = append(pixels, y)
	}
	return pixels
}

func (p *ProcessImage) TensorToImage(pixels [][]color.Color) (image.Image, error) {
	if len(pixels) == 0 {
		return nil, errors.New("tensor is empty")
	}
	rect := image.Rect(0, 0, len(pixels), len(pixels[0]))
	nImg := image.NewRGBA(rect)
	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[0]); y++ {
			q := pixels[x]
			if q == nil {
				continue
			}
			p := pixels[x][y]
			if p == nil {
				continue
			}
			original, ok := color.RGBAModel.Convert(p).(color.RGBA)
			if ok {
				nImg.Set(x, y, original)
			}
		}
	}

	return nImg, nil
}

func (p *ProcessImage) SaveImage(img *image.RGBA) error {
	f, err := os.Create(p.OutputPath)
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	if err != nil {
		return err
	}
	return f.Close()
}

func (p *ProcessImage) RotateImage90DegreesLeft(tensor [][]color.Color) [][]color.Color {
	if len(tensor) == 0 {
		return nil
	}
	rowCount := len(tensor)
	colCount := len(tensor[0])
	newTensor := make([][]color.Color, colCount)
	for i := range newTensor {
		newTensor[i] = make([]color.Color, rowCount)
	}

	for i := 0; i < rowCount; i++ {
		for j := 0; j < colCount; j++ {
			newTensor[j][rowCount-i-1] = tensor[i][j]
		}
	}

	return newTensor
}

func (p *ProcessImage) RotateImage90DegreesRight(tensor [][]color.Color) [][]color.Color {
	if len(tensor) == 0 {
		return nil
	}
	rowCount := len(tensor)
	colCount := len(tensor[0])
	newTensor := make([][]color.Color, colCount)
	for i := range newTensor {
		newTensor[i] = make([]color.Color, rowCount)
	}

	for i := 0; i < rowCount; i++ {
		for j := 0; j < colCount; j++ {
			newTensor[colCount-j-1][i] = tensor[i][j]
		}
	}

	return newTensor
}
