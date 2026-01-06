package img_storage

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// AddTextWatermark adds a text watermark to the image at the given path.
// Overwrites the original file with the watermarked image.
func AddTextWatermark(imagePath, text string) error {
	return AddTextWatermarkWithOpacity(imagePath, text, 0.5, color.RGBA{255, 255, 255, 255})
}

// AddTextWatermarkWithOpacity adds a text watermark with custom opacity and color.
func AddTextWatermarkWithOpacity(imagePath, text string, opacity float64, textColor color.Color) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image context
	dc := gg.NewContext(width, height)

	// Draw the original image
	dc.DrawImage(img, 0, 0)

	// Set font properties
	if err := dc.LoadFontFace("Arial.ttf", float64(width)/20); err != nil {
		// Fallback to default font if Arial is not available
		dc.SetFontFace(gg.NewFontFace(width/20))
	}

	// Set text color with opacity
	r, g, b, a := textColor.RGBA()
	textColor = color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * opacity),
	}
	dc.SetColor(textColor)

	// Measure text
	textWidth, textHeight := dc.MeasureString(text)
	textW := textWidth.Ceil()
	textH := textHeight.Ceil()

	// Position text at bottom-right corner with padding
	padding := width / 20
	x := float64(width - textW - padding)
	y := float64(height - textH - padding)

	// Draw rotated text (45 degrees)
	dc.RotateAbout(gg.Radians(-45), float64(width)/2, float64(height)/2)
	dc.DrawStringAnchored(text, float64(width)/2, float64(height)/2, 0.5, 0.5)

	// Get the final image
	watermarked := dc.Image()

	return saveImage(imagePath, watermarked, format)
}

// AddTextWatermarkCenter adds centered text watermark.
func AddTextWatermarkCenter(imagePath, text string, opacity float64, textColor color.Color) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dc := gg.NewContext(width, height)
	dc.DrawImage(img, 0, 0)

	// Set font size proportional to image
	fontSize := float64(width) / 15
	if err := dc.LoadFontFace("Arial.ttf", fontSize); err != nil {
		dc.SetFontFace(gg.NewFontFace(int(fontSize)))
	}

	r, g, b, a := textColor.RGBA()
	textColor = color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * opacity),
	}
	dc.SetColor(textColor)

	dc.DrawStringAnchored(text, float64(width)/2, float64(height)/2, 0.5, 0.5)

	watermarked := dc.Image()

	return saveImage(imagePath, watermarked, format)
}

// AddImageWatermark adds an image watermark to the target image.
// The watermark will be placed at the bottom-right corner.
func AddImageWatermark(imagePath, watermarkPath string, scale float64) error {
	return addImageWatermarkPosition(imagePath, watermarkPath, scale, "bottom-right")
}

// AddImageWatermarkCenter adds an image watermark at the center.
func AddImageWatermarkCenter(imagePath, watermarkPath string, scale float64) error {
	return addImageWatermarkPosition(imagePath, watermarkPath, scale, "center")
}

// AddImageWatermarkTopLeft adds an image watermark at the top-left corner.
func AddImageWatermarkTopLeft(imagePath, watermarkPath string, scale float64) error {
	return addImageWatermarkPosition(imagePath, watermarkPath, scale, "top-left")
}

// addImageWatermarkPosition adds an image watermark at the specified position.
// position can be: "center", "top-left", "top-right", "bottom-left", "bottom-right"
func addImageWatermarkPosition(imagePath, watermarkPath string, scale float64, position string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Load watermark image
	watermarkFile, err := os.Open(watermarkPath)
	if err != nil {
		return fmt.Errorf("failed to open watermark: %w", err)
	}
	defer watermarkFile.Close()

	watermark, _, err := image.Decode(watermarkFile)
	if err != nil {
		return fmt.Errorf("failed to decode watermark: %w", err)
	}

	// Scale watermark
	imgBounds := img.Bounds()
	watermarkBounds := watermark.Bounds()

	newWidth := int(float64(imgBounds.Dx()) * scale)
	newHeight := int(float64(newWidth) * float64(watermarkBounds.Dy()) / float64(watermarkBounds.Dx()))

	scaledWatermark := imaging.Resize(watermark, newWidth, newHeight, imaging.Lanczos)

	// Calculate position based on parameter
	var x, y int
	padding := 20
	switch position {
	case "center":
		x = (imgBounds.Dx() - newWidth) / 2
		y = (imgBounds.Dy() - newHeight) / 2
	case "top-left":
		x = padding
		y = padding
	case "top-right":
		x = imgBounds.Dx() - newWidth - padding
		y = padding
	case "bottom-left":
		x = padding
		y = imgBounds.Dy() - newHeight - padding
	case "bottom-right":
		x = imgBounds.Dx() - newWidth - padding
		y = imgBounds.Dy() - newHeight - padding
	default:
		x = imgBounds.Dx() - newWidth - padding
		y = imgBounds.Dy() - newHeight - padding
	}

	watermarked := imaging.Overlay(img, scaledWatermark, image.Pt(x, y), 1.0)

	return saveImage(imagePath, watermarked, format)
}

// AddRepeatedWatermark adds a repeated text watermark across the entire image.
func AddRepeatedWatermark(imagePath, text string, opacity float64, textColor color.Color) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	dc := gg.NewContext(width, height)
	dc.DrawImage(img, 0, 0)

	fontSize := float64(width) / 15
	if err := dc.LoadFontFace("Arial.ttf", fontSize); err != nil {
		dc.SetFontFace(gg.NewFontFace(int(fontSize)))
	}

	r, g, b, a := textColor.RGBA()
	textColor = color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * opacity),
	}
	dc.SetColor(textColor)

	// Calculate grid
	textWidth, _ := dc.MeasureString(text)
	textW := textWidth.Ceil()
	rows := height / (textW * 2)
	cols := width / (textW * 2)

	dc.RotateAbout(gg.Radians(-30), float64(width)/2, float64(height)/2)

	for i := -rows; i <= rows*2; i++ {
		for j := -cols; j <= cols*2; j++ {
			x := float64(j*textW*2) + float64(width)/2
			y := float64(i*textW) + float64(height)/2
			dc.DrawStringAnchored(text, x, y, 0.5, 0.5)
		}
	}

	watermarked := dc.Image()

	return saveImage(imagePath, watermarked, format)
}

// AddTiledWatermark adds a tiled image watermark across the entire image.
func AddTiledWatermark(imagePath, watermarkPath string, scale, opacity float64) error {
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	watermarkFile, err := os.Open(watermarkPath)
	if err != nil {
		return fmt.Errorf("failed to open watermark: %w", err)
	}
	defer watermarkFile.Close()

	watermark, _, err := image.Decode(watermarkFile)
	if err != nil {
		return fmt.Errorf("failed to decode watermark: %w", err)
	}

	imgBounds := img.Bounds()
	watermarkBounds := watermark.Bounds()

	tileWidth := int(float64(imgBounds.Dx()) * scale)
	tileHeight := int(float64(tileWidth) * float64(watermarkBounds.Dy()) / float64(watermarkBounds.Dx()))

	scaledWatermark := imaging.Resize(watermark, tileWidth, tileHeight, imaging.Lanczos)

	// Create tiled pattern
	result := imaging.Clone(img)
	for y := 0; y < imgBounds.Dy(); y += tileHeight {
		for x := 0; x < imgBounds.Dx(); x += tileWidth {
			result = imaging.Overlay(result, scaledWatermark, image.Pt(x, y), opacity)
		}
	}

	return saveImage(imagePath, result, format)
}
