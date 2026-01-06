package img_storage

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// createTestImage creates a simple test image at the specified path
func createTestImage(t *testing.T, path string, width, height int) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
}

// createTestWatermarkImage creates a simple watermark image
func createTestWatermarkImage(t *testing.T, path string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 100, 50))

	// Create a semi-transparent white rectangle
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 200})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create watermark image: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode watermark image: %v", err)
	}
}

// getFileInfo returns file info or nil if file doesn't exist
func getFileInfo(t *testing.T, path string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("failed to stat file: %v", err)
	}
	return info
}

func TestAddTextWatermark(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		wantError bool
	}{
		{
			name:      "simple text",
			text:      "Test Watermark",
			wantError: false,
		},
		{
			name:      "empty text",
			text:      "",
			wantError: false,
		},
		{
			name:      "special characters",
			text:      "© 2024 Test™",
			wantError: false,
		},
		{
			name:      "long text",
			text:      "This is a very long watermark text that should still be processed correctly",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			createTestImage(t, testImage, 800, 600)

			// Get original file info
			originalInfo := getFileInfo(t, testImage)

			// Execute
			err := AddTextWatermark(testImage, tt.text)

			// Verify
			if (err != nil) != tt.wantError {
				t.Errorf("AddTextWatermark() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				newInfo := getFileInfo(t, testImage)
				if newInfo == nil {
					t.Fatal("output file was not created")
				}
				if newInfo.Size() == originalInfo.Size() {
					t.Log("warning: file size unchanged, watermark may not have been applied")
				}
			}
		})
	}
}

func TestAddTextWatermarkWithOpacity(t *testing.T) {
	tests := []struct {
		name      string
		opacity   float64
		color     color.Color
		wantError bool
	}{
		{
			name:      "full opacity white",
			opacity:   1.0,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "half opacity white",
			opacity:   0.5,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "low opacity",
			opacity:   0.1,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "black text",
			opacity:   0.7,
			color:     color.RGBA{0, 0, 0, 255},
			wantError: false,
		},
		{
			name:      "red text",
			opacity:   0.8,
			color:     color.RGBA{255, 0, 0, 255},
			wantError: false,
		},
		{
			name:      "zero opacity",
			opacity:   0.0,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "invalid opacity negative",
			opacity:   -0.5,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false, // function doesn't validate opacity
		},
		{
			name:      "invalid opacity > 1",
			opacity:   1.5,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			createTestImage(t, testImage, 800, 600)

			err := AddTextWatermarkWithOpacity(testImage, "Test", tt.opacity, tt.color)

			if (err != nil) != tt.wantError {
				t.Errorf("AddTextWatermarkWithOpacity() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				if getFileInfo(t, testImage) == nil {
					t.Error("output file was not created")
				}
			}
		})
	}
}

func TestAddTextWatermarkCenter(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		opacity   float64
		color     color.Color
		wantError bool
	}{
		{
			name:      "centered white text",
			text:      "CENTERED",
			opacity:   0.8,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "centered black text",
			text:      "Centered Black",
			opacity:   0.9,
			color:     color.RGBA{0, 0, 0, 255},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			createTestImage(t, testImage, 800, 600)

			err := AddTextWatermarkCenter(testImage, tt.text, tt.opacity, tt.color)

			if (err != nil) != tt.wantError {
				t.Errorf("AddTextWatermarkCenter() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && getFileInfo(t, testImage) == nil {
				t.Error("output file was not created")
			}
		})
	}
}

func TestAddImageWatermark(t *testing.T) {
	tests := []struct {
		name      string
		scale     float64
		wantError bool
	}{
		{
			name:      "small scale",
			scale:     0.05,
			wantError: false,
		},
		{
			name:      "medium scale",
			scale:     0.1,
			wantError: false,
		},
		{
			name:      "large scale",
			scale:     0.3,
			wantError: false,
		},
		{
			name:      "very small scale",
			scale:     0.01,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			watermarkImage := filepath.Join(tmpDir, "watermark.png")

			createTestImage(t, testImage, 800, 600)
			createTestWatermarkImage(t, watermarkImage)

			err := AddImageWatermark(testImage, watermarkImage, tt.scale)

			if (err != nil) != tt.wantError {
				t.Errorf("AddImageWatermark() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && getFileInfo(t, testImage) == nil {
				t.Error("output file was not created")
			}
		})
	}
}

func TestAddImageWatermarkCenter(t *testing.T) {
	tests := []struct {
		name      string
		scale     float64
		wantError bool
	}{
		{
			name:      "centered small",
			scale:     0.1,
			wantError: false,
		},
		{
			name:      "centered medium",
			scale:     0.2,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			watermarkImage := filepath.Join(tmpDir, "watermark.png")

			createTestImage(t, testImage, 800, 600)
			createTestWatermarkImage(t, watermarkImage)

			err := AddImageWatermarkCenter(testImage, watermarkImage, tt.scale)

			if (err != nil) != tt.wantError {
				t.Errorf("AddImageWatermarkCenter() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && getFileInfo(t, testImage) == nil {
				t.Error("output file was not created")
			}
		})
	}
}

func TestAddImageWatermarkTopLeft(t *testing.T) {
	t.Run("top left position", func(t *testing.T) {
		tmpDir := t.TempDir()
		testImage := filepath.Join(tmpDir, "test.png")
		watermarkImage := filepath.Join(tmpDir, "watermark.png")

		createTestImage(t, testImage, 800, 600)
		createTestWatermarkImage(t, watermarkImage)

		err := AddImageWatermarkTopLeft(testImage, watermarkImage, 0.1)

		if err != nil {
			t.Errorf("AddImageWatermarkTopLeft() error = %v", err)
		}

		if getFileInfo(t, testImage) == nil {
			t.Error("output file was not created")
		}
	})
}

func TestAddRepeatedWatermark(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		opacity   float64
		color     color.Color
		wantError bool
	}{
		{
			name:      "repeated watermark",
			text:      "COPYRIGHT",
			opacity:   0.3,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "repeated low opacity",
			text:      "CONFIDENTIAL",
			opacity:   0.1,
			color:     color.RGBA{255, 255, 255, 255},
			wantError: false,
		},
		{
			name:      "repeated dark text",
			text:      "DRAFT",
			opacity:   0.5,
			color:     color.RGBA{0, 0, 0, 255},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			createTestImage(t, testImage, 800, 600)

			err := AddRepeatedWatermark(testImage, tt.text, tt.opacity, tt.color)

			if (err != nil) != tt.wantError {
				t.Errorf("AddRepeatedWatermark() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && getFileInfo(t, testImage) == nil {
				t.Error("output file was not created")
			}
		})
	}
}

func TestAddTiledWatermark(t *testing.T) {
	tests := []struct {
		name      string
		scale     float64
		opacity   float64
		wantError bool
	}{
		{
			name:      "tiled small",
			scale:     0.1,
			opacity:   0.5,
			wantError: false,
		},
		{
			name:      "tiled medium",
			scale:     0.15,
			opacity:   0.3,
			wantError: false,
		},
		{
			name:      "tiled low opacity",
			scale:     0.1,
			opacity:   0.1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testImage := filepath.Join(tmpDir, "test.png")
			watermarkImage := filepath.Join(tmpDir, "watermark.png")

			createTestImage(t, testImage, 800, 600)
			createTestWatermarkImage(t, watermarkImage)

			err := AddTiledWatermark(testImage, watermarkImage, tt.scale, tt.opacity)

			if (err != nil) != tt.wantError {
				t.Errorf("AddTiledWatermark() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && getFileInfo(t, testImage) == nil {
				t.Error("output file was not created")
			}
		})
	}
}

func TestAddImageWatermarkErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		imagePath     string
		watermarkPath string
		scale         float64
		wantError     bool
	}{
		{
			name:          "missing image file",
			imagePath:     "nonexistent.png",
			watermarkPath: "watermark.png",
			scale:         0.1,
			wantError:     true,
		},
		{
			name:          "missing watermark file",
			imagePath:     "image.png",
			watermarkPath: "nonexistent.png",
			scale:         0.1,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.imagePath == "image.png" {
				tt.imagePath = filepath.Join(tmpDir, "image.png")
				createTestImage(t, tt.imagePath, 800, 600)
			} else {
				tt.imagePath = filepath.Join(tmpDir, tt.imagePath)
			}

			if tt.watermarkPath == "watermark.png" {
				tt.watermarkPath = filepath.Join(tmpDir, "watermark.png")
				createTestWatermarkImage(t, tt.watermarkPath)
			} else {
				tt.watermarkPath = filepath.Join(tmpDir, tt.watermarkPath)
			}

			err := AddImageWatermark(tt.imagePath, tt.watermarkPath, tt.scale)

			if (err != nil) != tt.wantError {
				t.Errorf("AddImageWatermark() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAddTextWatermarkErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		imagePath string
		wantError bool
	}{
		{
			name:      "missing image file",
			imagePath: "nonexistent.png",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tt.imagePath = filepath.Join(tmpDir, tt.imagePath)

			err := AddTextWatermark(tt.imagePath, "Test")

			if (err != nil) != tt.wantError {
				t.Errorf("AddTextWatermark() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// Benchmark tests
func BenchmarkAddTextWatermark(b *testing.B) {
	tmpDir := b.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	createTestImage(&testing.T{}, testImage, 1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := AddTextWatermark(testImage, "Benchmark"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddImageWatermark(b *testing.B) {
	tmpDir := b.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	watermarkImage := filepath.Join(tmpDir, "watermark.png")
	createTestImage(&testing.T{}, testImage, 1920, 1080)
	createTestWatermarkImage(&testing.T{}, watermarkImage)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := AddImageWatermark(testImage, watermarkImage, 0.1); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddRepeatedWatermark(b *testing.B) {
	tmpDir := b.TempDir()
	testImage := filepath.Join(tmpDir, "test.png")
	createTestImage(&testing.T{}, testImage, 1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := AddRepeatedWatermark(testImage, "COPYRIGHT", 0.3, color.RGBA{255, 255, 255, 255}); err != nil {
			b.Fatal(err)
		}
	}
}
