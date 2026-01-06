package img_storage

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// TODO: implement next update image functions:
// TODO: implement resize function
// TODO: implement miniature generator
// TODO: implement watermark creating function

// Resize resizes the image at the given path to the specified width and height.
// Overwrites the original file with the resized image.
func Resize(imagePath string, width, height int) error {
	// Open the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize the image using Lanczos resampling
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Determine output format
	outputFormat := format
	if outputFormat == "" {
		ext := strings.ToLower(filepath.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			outputFormat = "jpeg"
		case ".png":
			outputFormat = "png"
		case ".gif":
			outputFormat = "gif"
		default:
			outputFormat = "png"
		}
	}

	// Create a temporary file to save the resized image
	tmpPath := imagePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Encode and save to temp file
	var encodeErr error
	switch outputFormat {
	case "jpeg":
		encodeErr = imaging.Encode(tmpFile, resized, imaging.JPEG, imaging.JPEGQuality(85))
	case "png":
		encodeErr = imaging.Encode(tmpFile, resized, imaging.PNG, imaging.PNGCompressionLevel(9))
	case "gif":
		encodeErr = imaging.Encode(tmpFile, resized, imaging.GIF)
	default:
		encodeErr = imaging.Encode(tmpFile, resized, imaging.PNG, imaging.JPEGQuality(85))
	}
	tmpFile.Close()

	if encodeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode image: %w", encodeErr)
	}

	// Replace original file with resized one
	if err := os.Rename(tmpPath, imagePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace original file: %w", err)
	}

	return nil
}

// ResizeToFit resizes the image to fit within the given dimensions while preserving aspect ratio.
func ResizeToFit(imagePath string, maxWidth, maxHeight int) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize to fit while preserving aspect ratio
	resized := imaging.Fit(img, maxWidth, maxHeight, imaging.Lanczos)

	// Save the resized image
	return saveImage(imagePath, resized, format)
}

// ResizeByWidth resizes the image to the specified width, preserving aspect ratio.
func ResizeByWidth(imagePath string, width int) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resized := imaging.Resize(img, width, 0, imaging.Lanczos)

	return saveImage(imagePath, resized, format)
}

// ResizeByHeight resizes the image to the specified height, preserving aspect ratio.
func ResizeByHeight(imagePath string, height int) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	resized := imaging.Resize(img, 0, height, imaging.Lanczos)

	return saveImage(imagePath, resized, format)
}

// saveImage saves the image to the specified path with the given format.
func saveImage(imagePath string, img *image.NRGBA, format string) error {
	tmpPath := imagePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	var encodeErr error
	switch format {
	case "jpeg":
		encodeErr = imaging.Encode(tmpFile, img, imaging.JPEG, imaging.JPEGQuality(85))
	case "png":
		encodeErr = imaging.Encode(tmpFile, img, imaging.PNG, imaging.PNGCompressionLevel(9))
	case "gif":
		encodeErr = imaging.Encode(tmpFile, img, imaging.GIF)
	default:
		encodeErr = imaging.Encode(tmpFile, img, imaging.JPEG, imaging.JPEGQuality(85))
	}
	tmpFile.Close()

	if encodeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode image: %w", encodeErr)
	}

	if err := os.Rename(tmpPath, imagePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace original file: %w", err)
	}

	return nil
}
