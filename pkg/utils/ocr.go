package utils

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/otiai10/gosseract/v2"
	"github.com/zsmartex/pkg/v2/log"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func nuke(f *os.File) {
	name := f.Name()
	f.Close()
	if err := os.Remove(name); err != nil {
		log.Error(err)
	}
}

// GetTextFromImage 使用 gosseract 进行 OCR 识别，直接从内存中的图像提取文本
func GetTextFromImage(img image.Image) (string, error) {
	// 创建一个新的 Tesseract 客户端
	client := gosseract.NewClient()
	defer client.Close()

	// 将 img 转换为 PNG 格式的字节数组
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	// 将图像字节数据传递给 Tesseract 客户端
	client.SetImageFromBytes(buf.Bytes())
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE) // 设置 PSM 模式，例如单个文本块
	// 执行 OCR 并返回结果
	text, err := client.Text()
	if err != nil {
		return "", fmt.Errorf("OCR failed: %v", err)
	}
	return text, nil
}

func createTempFile() (*os.File, error) {
	filename := fmt.Sprintf("%s.png", uuid.New())
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	tempDir := filepath.Join(".", "data")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return nil, err
	}
	return os.CreateTemp(tempDir, prefix+"_*"+ext)
}
