package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/otiai10/gosseract/v2"
)

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

	// Read and decode image
	img, format, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	log.Printf("Original image: %dx%d, format: %s", img.Bounds().Dx(), img.Bounds().Dy(), format)

	// Save original image
	originalBytes, _ := imageToBytes(img, "png")
	saveImageToFolder(originalBytes, "original_"+header.Filename)

	// Preprocess image for OCR
	processedImg := preprocessImageForOCR(img)

	// Save processed image for debugging
	processedBytes, _ := imageToBytes(processedImg, "png")
	saveImageToFolder(processedBytes, "processed_"+header.Filename)

	// Extract text using row-by-row processing
	extractedText := extractTextByRows(processedImg)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"data": %q}`, extractedText)))
}

func preprocessImageForOCR(img image.Image) image.Image {
	// Step 1: Resize for optimal OCR (height should be 300-600px for best results)
	bounds := img.Bounds()
	targetHeight := 500
	aspectRatio := float64(bounds.Dx()) / float64(bounds.Dy())
	targetWidth := int(float64(targetHeight) * aspectRatio)

	resized := imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)
	log.Printf("Resized to: %dx%d", targetWidth, targetHeight)

	// Step 2: Convert to grayscale
	grayscale := imaging.Grayscale(resized)

	// Step 3: Enhance contrast
	enhanced := imaging.AdjustContrast(grayscale, 30)
	enhanced = imaging.AdjustBrightness(enhanced, 10)

	// Step 4: Apply Gaussian blur to reduce noise
	blurred := imaging.Blur(enhanced, 0.5)

	// Step 5: Sharpen the image
	sharpened := imaging.Sharpen(blurred, 2.0)

	// Step 6: Apply threshold to create pure black and white
	threshold := applyAdaptiveThreshold(sharpened)

	// Step 7: Morphological operations to clean up
	cleaned := morphologicalCleaning(threshold)

	// Step 8: Edge enhancement
	edgeEnhanced := enhanceEdges(cleaned)

	return edgeEnhanced
}

func applyAdaptiveThreshold(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	// Convert to grayscale if not already
	grayImg := imaging.Grayscale(img)

	// Apply Otsu's threshold method
	histogram := make([]int, 256)

	// Build histogram
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(grayImg.At(x, y)).(color.Gray)
			histogram[gray.Y]++
		}
	}

	// Calculate optimal threshold using Otsu's method
	threshold := calculateOtsuThreshold(histogram, bounds.Dx()*bounds.Dy())
	log.Printf("Calculated threshold: %d", threshold)

	// Apply threshold
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(grayImg.At(x, y)).(color.Gray)
			if gray.Y > threshold {
				result.SetGray(x, y, color.Gray{255}) // White
			} else {
				result.SetGray(x, y, color.Gray{0}) // Black
			}
		}
	}

	return result
}

func calculateOtsuThreshold(histogram []int, totalPixels int) uint8 {
	var sum, sumB float64
	var wB, wF int
	var max, between, threshold float64

	// Calculate total sum
	for i := 0; i < 256; i++ {
		sum += float64(i * histogram[i])
	}

	for i := 0; i < 256; i++ {
		wB += histogram[i]
		if wB == 0 {
			continue
		}

		wF = totalPixels - wB
		if wF == 0 {
			break
		}

		sumB += float64(i * histogram[i])
		mB := sumB / float64(wB)
		mF := (sum - sumB) / float64(wF)

		between = float64(wB) * float64(wF) * (mB - mF) * (mB - mF)

		if between > max {
			max = between
			threshold = float64(i)
		}
	}

	return uint8(threshold)
}

func morphologicalCleaning(img image.Image) image.Image {
	// Simple morphological operations to remove noise
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	// Erosion followed by dilation (opening) to remove noise
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			// Check 3x3 neighborhood
			minVal := uint8(255)
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					gray := color.GrayModel.Convert(img.At(x+dx, y+dy)).(color.Gray)
					if gray.Y < minVal {
						minVal = gray.Y
					}
				}
			}
			result.SetGray(x, y, color.Gray{minVal})
		}
	}

	return result
}

func enhanceEdges(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewGray(bounds)

	// Sobel edge detection
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			// Get surrounding pixels
			var pixels [3][3]uint8
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					gray := color.GrayModel.Convert(img.At(x+dx, y+dy)).(color.Gray)
					pixels[dy+1][dx+1] = gray.Y
				}
			}

			// Sobel X kernel
			gx := -1*int(pixels[0][0]) + 1*int(pixels[0][2]) +
				-2*int(pixels[1][0]) + 2*int(pixels[1][2]) +
				-1*int(pixels[2][0]) + 1*int(pixels[2][2])

			// Sobel Y kernel
			gy := -1*int(pixels[0][0]) - 2*int(pixels[0][1]) - 1*int(pixels[0][2]) +
				1*int(pixels[2][0]) + 2*int(pixels[2][1]) + 1*int(pixels[2][2])

			// Combine gradients
			magnitude := math.Sqrt(float64(gx*gx + gy*gy))

			// Enhance edges while preserving text
			originalGray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			enhanced := float64(originalGray.Y) + magnitude*0.3

			if enhanced > 255 {
				enhanced = 255
			}

			result.SetGray(x, y, color.Gray{uint8(enhanced)})
		}
	}

	return result
}

func extractTextByRows(img image.Image) string {
	bounds := img.Bounds()
	height := bounds.Dy()

	// Analyze horizontal projections to find text rows
	rowProjections := make([]int, height)

	for y := 0; y < height; y++ {
		blackPixels := 0
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if gray.Y < 128 { // Count dark pixels
				blackPixels++
			}
		}
		rowProjections[y] = blackPixels
	}

	// Find text rows based on projection peaks
	textRows := findTextRows(rowProjections)
	log.Printf("Found %d text rows", len(textRows))

	var extractedTexts []string

	for i, row := range textRows {
		// Extract row image
		rowImg := imaging.Crop(img, image.Rect(bounds.Min.X, row.start, bounds.Max.X, row.end))

		// Save row for debugging
		rowBytes, _ := imageToBytes(rowImg, "png")
		saveImageToFolder(rowBytes, fmt.Sprintf("row_%d.png", i))

		// OCR this specific row
		text := performOCROnImage(rowImg)
		if strings.TrimSpace(text) != "" {
			extractedTexts = append(extractedTexts, strings.TrimSpace(text))
		}
	}

	return strings.Join(extractedTexts, "\n")
}

type TextRow struct {
	start, end int
	density    int
}

func findTextRows(projections []int) []TextRow {
	var rows []TextRow

	// Calculate average and threshold
	total := 0
	for _, p := range projections {
		total += p
	}
	avgProjection := total / len(projections)
	threshold := avgProjection / 3 // Adjust this value based on your images

	log.Printf("Average projection: %d, threshold: %d", avgProjection, threshold)

	inRow := false
	rowStart := 0

	for i, projection := range projections {
		if !inRow && projection > threshold {
			// Start of a text row
			inRow = true
			rowStart = i
		} else if inRow && projection <= threshold {
			// End of a text row
			inRow = false
			if i-rowStart > 10 { // Minimum row height
				rows = append(rows, TextRow{
					start:   rowStart,
					end:     i,
					density: calculateRowDensity(projections[rowStart:i]),
				})
			}
		}
	}

	// Handle case where last row extends to end
	if inRow {
		rows = append(rows, TextRow{
			start:   rowStart,
			end:     len(projections),
			density: calculateRowDensity(projections[rowStart:]),
		})
	}

	// Sort by density (most text first)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].density > rows[j].density
	})

	// Re-sort by position for proper reading order
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].start < rows[j].start
	})

	return rows
}

func calculateRowDensity(projections []int) int {
	total := 0
	for _, p := range projections {
		total += p
	}
	if len(projections) == 0 {
		return 0
	}
	return total / len(projections)
}

func performOCROnImage(img image.Image) string {
	// Convert image to bytes
	imgBytes, err := imageToBytes(img, "png")
	if err != nil {
		log.Printf("Failed to convert image to bytes: %v", err)
		return ""
	}

	client := gosseract.NewClient()
	defer client.Close()

	// Optimized settings for preprocessed images
	client.SetLanguage("eng")
	client.SetPageSegMode(7) // Single text line

	// Character whitelist for receipts
	client.SetVariable("tessedit_char_whitelist", "0123456789.,€$£¥-+ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz()[]{}:;/\\*&%@#!?'\"|`~^_= ")

	client.SetImageFromBytes(imgBytes)

	text, err := client.Text()
	if err != nil {
		log.Printf("OCR error on row: %v", err)
		return ""
	}

	return text
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

func saveImageToFolder(imageData []byte, filename string) error {
	uploadsDir := "./debug_images"
	err := os.MkdirAll(uploadsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create debug directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(uploadsDir, fmt.Sprintf("%s_%s", timestamp, filename))

	err = os.WriteFile(filePath, imageData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("Debug image saved: %s", filePath)
	return nil
}
