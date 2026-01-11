package img_storage

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// WatermarkConfig конфигурация для нанесения водяного знака
type WatermarkConfig struct {
	Text      string      // текст водяного знака
	FontSize  float64     // размер шрифта (0 = автоматический, пропорционально изображению)
	Opacity   float64     // прозрачность 0.0-1.0
	Color     color.Color // цвет текста
	Rotation  float64     // угол поворота в градусах
	PositionX string      // "left", "center", "right"
	PositionY string      // "top", "center", "bottom"
	OffsetX   int         // дополнительное смещение по X
	OffsetY   int         // дополнительное смещение по Y
}

// DefaultWatermarkConfig возвращает конфигурацию по умолчанию
func DefaultWatermarkConfig() *WatermarkConfig {
	return &WatermarkConfig{
		Text:      "watermark",
		FontSize:  0, // автоматический размер
		Opacity:   0.5,
		Color:     color.RGBA{255, 255, 255, 255},
		Rotation:  -45,
		PositionX: "center",
		PositionY: "center",
		OffsetX:   0,
		OffsetY:   0,
	}
}

// ApplyWatermark наносит водяной знак на изображение согласно конфигурации.
// Если config = nil, используются настройки по умолчанию.
func ApplyWatermark(imagePath string, config *WatermarkConfig) error {
	// Используем дефолтную конфигурацию если не передана
	if config == nil {
		config = DefaultWatermarkConfig()
	}

	// Открываем изображение
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть изображение: %w", err)
	}
	defer file.Close()

	// Декодируем изображение
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("не удалось декодировать изображение: %w", err)
	}

	// Получаем размеры изображения
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Создаём контекст для рисования
	dc := gg.NewContext(width, height)
	dc.DrawImage(img, 0, 0)

	// Устанавливаем размер шрифта
	fontSize := config.FontSize
	if fontSize == 0 {
		fontSize = float64(width) / 15 // автоматический размер
	}

	// Загружаем шрифт (используем дефолтный если Arial недоступен)
	//if err := dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", fontSize); err != nil {
	//	// Fallback на встроенный шрифт
	//	basicFont := gg.NewFontFace(int(fontSize))
	//	dc.SetFontFace(basicFont)
	//}

	// Устанавливаем цвет с учётом прозрачности
	r, g, b, a := config.Color.RGBA()
	textColor := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(float64(a>>8) * config.Opacity),
	}
	dc.SetColor(textColor)

	// Вычисляем позицию текста
	textWidth, textHeight := dc.MeasureString(config.Text)

	var x, y float64

	// Позиция по X
	switch config.PositionX {
	case "left":
		x = textWidth / 2
	case "right":
		x = float64(width) - textWidth/2
	default: // "center"
		x = float64(width) / 2
	}

	// Позиция по Y
	switch config.PositionY {
	case "top":
		y = textHeight
	case "bottom":
		y = float64(height) - textHeight
	default: // "center"
		y = float64(height) / 2
	}

	// Применяем смещение
	x += float64(config.OffsetX)
	y += float64(config.OffsetY)

	// Поворачиваем и рисуем текст
	if config.Rotation != 0 {
		dc.RotateAbout(gg.Radians(config.Rotation), x, y)
	}
	dc.DrawStringAnchored(config.Text, x, y, 0.5, 0.5)

	// Получаем итоговое изображение
	watermarked := dc.Image()

	// Сохраняем изображение
	return saveWatermarkedImage(imagePath, watermarked, format)
}

// saveWatermarkedImage сохраняет изображение с водяным знаком
func saveWatermarkedImage(imagePath string, img image.Image, format string) error {
	// Определяем формат если не указан
	if format == "" {
		ext := strings.ToLower(filepath.Ext(imagePath))
		switch ext {
		case ".jpg", ".jpeg":
			format = "jpeg"
		case ".png":
			format = "png"
		case ".gif":
			format = "gif"
		default:
			format = "png"
		}
	}

	// Создаём временный файл
	tmpPath := imagePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("не удалось создать временный файл: %w", err)
	}

	// Кодируем изображение
	var encodeErr error
	switch format {
	case "jpeg":
		encodeErr = imaging.Encode(tmpFile, img, imaging.JPEG, imaging.JPEGQuality(95))
	case "png":
		encodeErr = imaging.Encode(tmpFile, img, imaging.PNG)
	case "gif":
		encodeErr = imaging.Encode(tmpFile, img, imaging.GIF)
	default:
		encodeErr = imaging.Encode(tmpFile, img, imaging.PNG)
	}
	tmpFile.Close()

	if encodeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось закодировать изображение: %w", encodeErr)
	}

	// Заменяем оригинальный файл
	if err := os.Rename(tmpPath, imagePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("не удалось заменить оригинальный файл: %w", err)
	}

	return nil
}
