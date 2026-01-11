package consumer

import (
	"encoding/json"
	"fmt"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/models"
	"imageProcessor/internal/storage/sqlite"
)

// offered actions with images
var actions = map[string]bool{
	resizeAction:    true,
	miniatureAction: true,
	watermarkAction: true,
}

var tmpResizeParameters = [2]int{20, 25}

const (
	resizeAction    = "resize"
	miniatureAction = "miniature"
	watermarkAction = "watermark"
)

// ConsumedHandler is designed to handle fetched messages
func ConsumedHandler(msg []byte, storage *sqlite.StorageSqlite) error {
	const op = "kafka.consumer.ConsumerHandler"
	if len(msg) == 0 {
		return fmt.Errorf("message is empty")
	}

	var kafkaMessage models.KafkaMessage
	err := json.Unmarshal(msg, &kafkaMessage)
	if err != nil {
		return fmt.Errorf("unmarshaled message error; %s, %w", op, err)
	}

	if _, ok := actions[kafkaMessage.Action]; !ok {
		return fmt.Errorf("incorrect recived action; %s", op)
	}

	metadata, err := storage.GetImageMetadata(kafkaMessage.Id)
	if err != nil {
		return fmt.Errorf("%s,%w", op, err)
	}

	switch kafkaMessage.Action {
	case resizeAction:
		err := img_storage.ResizeImage(metadata.OriginalPath, tmpResizeParameters[0], tmpResizeParameters[1])
		if err != nil {
			return fmt.Errorf("%s, %w", op, err)
		}
	case miniatureAction:
		// TODO: add miniature function itself
		// WARN: temporarily ise ResizeImage because this function has the same approach
		err := img_storage.ResizeImage(metadata.OriginalPath, tmpResizeParameters[0], tmpResizeParameters[1])
		if err != nil {
			return fmt.Errorf("%s, %w", op, err)
		}
	case watermarkAction:
		err := img_storage.ApplyWatermark(metadata.OriginalPath, img_storage.DefaultWatermarkConfig())
		if err != nil {
			return fmt.Errorf("%s, %w", op, err)
		}
	default:
		return fmt.Errorf("incorrect action; %s, %w", op, err)
	}

	return nil

}
