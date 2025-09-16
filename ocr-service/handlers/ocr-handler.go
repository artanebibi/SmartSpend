package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	_ "math"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/jdeng/goheif"
	"github.com/otiai10/gosseract/v2"
)

type OCRResult struct {
	Version     string  `json:"version"`
	Text        string  `json:"text"`
	Confidence  float64 `json:"confidence,omitempty"`
	Description string  `json:"description"`
}

type OCRResponse struct {
	Results []OCRResult `json:"results"`
}

func OcrHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	var img image.Image

	if ext == ".heic" || ext == ".heif" {
		img, err = goheif.Decode(file)
		if err != nil {
			http.Error(w, "Failed to decode HEIC image", http.StatusBadRequest)
			return
		}
	} else {
		img, _, err = image.Decode(file)
		if err != nil {
			http.Error(w, "Failed to decode image", http.StatusBadRequest)
			return
		}
	}

	results := []OCRResult{
		//performOCRBasic(img),
		performOCREnhanced(img),
		//performOCRHighContrast(img),
		performOCRDilated(img),
		performOCRWithDenoising(img),
		//performOCRScaled(img),
	}

	response := OCRResponse{Results: results}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func performOCRBasic(img image.Image) OCRResult {
	grayscale := imaging.Grayscale(img)
	sharpened := imaging.Sharpen(grayscale, 2.0)

	text := performOCRWithConfig(sharpened, "mkd", map[string]string{})

	return OCRResult{
		Version:     "basic",
		Text:        text,
		Description: "Basic grayscale + sharpen processing",
	}
}

func performOCREnhanced(img image.Image) OCRResult {
	grayscale := imaging.Grayscale(img)
	contrasted := imaging.AdjustContrast(grayscale, 20)
	sharpened := imaging.Sharpen(contrasted, 1.5)
	brightened := imaging.AdjustBrightness(sharpened, 5)

	text := performOCRWithConfig(brightened, "mkd", map[string]string{})

	return OCRResult{
		Version:     "enhanced",
		Text:        text,
		Description: "Enhanced contrast + brightness + character whitelist",
	}
}

func performOCRHighContrast(img image.Image) OCRResult {
	grayscale := imaging.Grayscale(img)
	thresholded := applyThreshold(grayscale, 0.6)

	text := performOCRWithConfig(thresholded, "eng+mkd", map[string]string{})

	return OCRResult{
		Version:     "high_contrast",
		Text:        text,
		Description: "High contrast thresholding",
	}
}

func performOCRDilated(img image.Image) OCRResult {
	grayscale := imaging.Grayscale(img)
	contrasted := imaging.AdjustContrast(grayscale, 15)
	dilated := applyDilation(contrasted, 1)

	text := performOCRWithConfig(dilated, "mkd", map[string]string{})

	return OCRResult{
		Version:     "dilated",
		Text:        text,
		Description: "Contrast + morphological dilation",
	}
}

func performOCRWithDenoising(img image.Image) OCRResult {
	grayscale := imaging.Grayscale(img)

	denoised := imaging.Blur(grayscale, 0.5)

	sharpened := imaging.Sharpen(denoised, 3.0)

	contrasted := imaging.AdjustContrast(sharpened, 25)

	text := performOCRWithConfig(contrasted, "mkd", map[string]string{})

	return OCRResult{
		Version:     "denoised",
		Text:        text,
		Description: "Gaussian blur denoising + aggressive sharpening",
	}
}

func performOCRScaled(img image.Image) OCRResult {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	scaled := imaging.Resize(img, width*2, height*2, imaging.Lanczos)

	grayscale := imaging.Grayscale(scaled)
	sharpened := imaging.Sharpen(grayscale, 1.2)

	text := performOCRWithConfig(sharpened, "mkd", map[string]string{})

	return OCRResult{
		Version:     "scaled_2x",
		Text:        text,
		Description: "2x upscaled with Lanczos resampling",
	}
}

func performOCRWithConfig(img image.Image, language string, config map[string]string) string {
	imgBytes, err := imageToBytes(img, "png")
	if err != nil {
		log.Printf("Failed to convert image to bytes: %v", err)
		return ""
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.SetLanguage("mkd")
	client.SetPageSegMode(gosseract.PSM_SINGLE_BLOCK)
	client.SetVariable("tessedit_char_whitelist",
		"АБВГДЕЖЗИЈКЛЉМНЊОПРСТЌУФХЦЧЏШабвгдежзијклљмнњопрстќуфхцчџш0123456789.,:/-%")
	for key, value := range config {
		client.SetVariable(gosseract.SettableVariable(key), value)
	}

	client.SetImageFromBytes(imgBytes)

	text, err := client.Text()
	if err != nil {
		log.Printf("OCR error: %v", err)
		return ""
	}

	return strings.TrimSpace(text)
}

func applyThreshold(img image.Image, threshold float64) image.Image {
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			normalized := float64(grayColor.Y) / 255.0

			var newValue uint8
			if normalized > threshold {
				newValue = 255
			} else {
				newValue = 0
			}

			result.SetGray(x, y, color.Gray{Y: newValue})
		}
	}

	return result
}

// Simple dilation operation
func applyDilation(img image.Image, radius int) image.Image {
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			maxValue := uint8(0)

			// Check neighborhood
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						grayColor := color.GrayModel.Convert(img.At(nx, ny)).(color.Gray)
						if grayColor.Y > maxValue {
							maxValue = grayColor.Y
						}
					}
				}
			}

			result.SetGray(x, y, color.Gray{Y: maxValue})
		}
	}

	return result
}

func imageToBytes(img image.Image, format string) ([]byte, error) {
	buf := new(bytes.Buffer)

	switch format {
	case "png":
		err := png.Encode(buf, img)
		return buf.Bytes(), err
	case "jpeg", "jpg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 95})
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
