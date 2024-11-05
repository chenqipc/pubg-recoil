package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

func CropImage(img image.Image, bounds image.Rectangle) (image.Image, error) {
	return cutter.Crop(img, cutter.Config{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Anchor: image.Point{
			X: bounds.Min.X,
			Y: bounds.Min.Y,
		},
	})
}

func GetPixel(img image.Image, x, y int) color.Color {
	return img.At(x, y)
}

func ImageToGrayScale(img image.Image) *image.Gray {
	result := image.NewGray(img.Bounds())
	draw.Draw(result, result.Bounds(), img, img.Bounds().Min, draw.Src)

	return result
}

func ImageResize(img image.Image, scale int) image.Image {
	return resize.Resize(uint(img.Bounds().Dx()*scale), uint(img.Bounds().Dy()*scale), img, resize.Lanczos3)
}

// SaveImageToFile 保存图像到本地文件的辅助函数
func SaveImageToFile(img image.Image, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	return nil
}
