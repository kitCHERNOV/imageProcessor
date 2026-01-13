package consumer

import (
	"encoding/json"
	"fmt"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/models"
	"imageProcessor/internal/storage/sqlite"
	"log/slog"
	"os"
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

// Statuses
const (
	processingStatus = "processing"
	modifiedStatus   = "modified"
	failedStatus     = "failed"
)

// ConsumedHandler is designed to handle fetched messages
func ConsumedHandler(msg []byte, storage *sqlite.StorageSqlite, log *slog.Logger) error {
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
	log.Debug("request action is checked", "action", kafkaMessage.Action)

	metadata, err := storage.GetImageMetadata(kafkaMessage.Id)
	if err != nil {
		return fmt.Errorf("%s,%w", op, err)
	}

	log.Debug("consumer receive metadata by id", slog.Int("id", kafkaMessage.Id))
	log.Debug("image is turn processing")

	if metadata.Status == "deleted" {
		// deleting the image itself
		err = os.Remove(metadata.OriginalPath)
		if err != nil {
			log.Error("Removing original file failed", "op", op, "err", err)
			//w.WriteHeader(http.StatusInternalServerError)
			//w.Write([]byte("file delete failed"))
			return fmt.Errorf("%s,%w", op, err)
		}
		// deleting the image metadata
		err = storage.DeleteImage(kafkaMessage.Id)
		if err != nil {
			log.Error("metadata deleting error", "op", op, "err", err)
			//w.WriteHeader(http.StatusInternalServerError)
			//w.Write([]byte("db delete failed"))
			return fmt.Errorf("%s,%w", op, err)
		}
		return nil
	}

	// TODO: add logic to change status
	// TODO: add worker pool

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
	// after updating change status parameter
	err = storage.UpdateStatus(kafkaMessage.Id, modifiedStatus)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}
	log.Debug("image is modified",
		slog.Int("Id", kafkaMessage.Id),
		slog.String("action", kafkaMessage.Action),
	)

	return nil

}
