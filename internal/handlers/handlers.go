package handlers

import (
	"encoding/json"
	"fmt"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/kafka"
	"imageProcessor/internal/models"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type ImageSqlSaver interface {
	SetMetadata(metadata *models.ImageMetadata) (int, error)
	DownloadImage(id int) (*models.ImageMetadata, error)
	DeleteImage(id int) error
}

type ImageActionRequest struct {
	Image  string `json:"image"`
	Action string `json:"action"`
}

// ImageActionResponse - структура для ответа со статусом 202 Accepted
type ImageActionResponse struct {
	Status  string `json:"status"`   // "accepted", "processing", "completed"
	Message string `json:"message"`  // описание результата
	ImageID int    `json:"image_id"` // уникальный ID картинки
	Action  string `json:"action"`   // выполняемое действие
	//TaskID      string     `json:"task_id"`                // ID асинхронной задачи
	//CreatedAt   time.Time  `json:"created_at"`             // время создания запроса
	//CompletedAt *time.Time `json:"completed_at,omitempty"` // время завершения
}

// ImageIncludedResponse - struct to send an image in response to client
type ImageIncludedResponse struct {
	Status  string `json:"status"`
	Image   []byte `json:"image"`
	Message string `json:"message"`
}

func (r *ImageActionRequest) RequestValidate() error {
	if r.Image == "" {
		return toWrapHandlersErrors(ErrEmptyImage)
	}
	if r.Action == "" {
		return toWrapHandlersErrors(ErrEmptyAction)
	}
	return nil
}

// constants
const (
	imgForm    = "image"
	actionForm = "action"
)

// statuses of image handling
const (
	modifiedStatus    = "modified"
	resizedStatus     = "image was resized"
	miniaturedStatus  = "miniature was created"
	watermarkedStatus = "watermark wad added"
)

func UploadImage(log *slog.Logger, storage ImageSqlSaver, imgStorage img_storage.ImageStorage, producer kafka.Producer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.UploadImage"

		// check byte form size
		var maxMemory int64 = 10 * 1024 * 1024
		if err := r.ParseMultipartForm(maxMemory); err != nil {
			// Если форма больше 10MB, будет ошибка
			http.Error(w, "Gotten data too large", http.StatusRequestEntityTooLarge)
			return
		}

		// TODO: implement fork of some approaches (JSON and Form)

		// Form approach temporarily here
		file, handler, err := r.FormFile("image")
		if err != nil {
			log.Error("error getting file", "op", op, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		// check extension
		extension := filepath.Ext(handler.Filename)

		allowedExtensions := map[string]bool{
			".jpg": true,
			".png": true,
			".gif": true,
		}

		if !allowedExtensions[extension] {
			log.Warn("%s; file extension not allowed", op)
			http.Error(w, fmt.Sprintf("invalid file extension"), http.StatusBadRequest)
			return
		}
		// get action parameter

		action := r.FormValue("action")

		// upload file into local storage
		baseFilename := filepath.Base(handler.Filename)
		newFilePath := filepath.Join(imgStorage.ImgStoragePath, baseFilename)
		dst, err := os.Create(newFilePath)
		if err != nil {
			log.Error("%s; error creating file: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			log.Error("%s; error uploading file: %v", op, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// TODO: to add sqlite SetMetaData function
		imgMetadata := models.ImageMetadata{
			OriginalFilename: baseFilename,
			OriginalPath:     newFilePath,
			MimeType:         extension,
			FileSize:         int(handler.Size),
			Status:           "pending",
			Action:           action,
		}

		id, err := storage.SetMetadata(&imgMetadata)
		if err != nil {
			log.Error("Adding new image's metadata failed", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		//TODO: to add kafka producer calling
		kafkaMessage := models.KafkaMessage{
			Id:     id,
			Action: action,
		}
		// TODO: To add topic into configuration file
		const tempTopic = "gotten_task"
		err = producer.SendMessage(tempTopic, kafkaMessage)
		if err != nil {
			log.Error("sending message into broker failed", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}

		// create a response message
		response := ImageActionResponse{
			Status:  http.StatusText(http.StatusAccepted),
			Message: fmt.Sprintf("image is uploaded seccessuly to do - %s", action),
			ImageID: id,
			Action:  action,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
		//fmt.Fprintf(w, "File %s downloaded seccessfuly", handler.Filename)
	}
}

const idQueryParameter = "id"

func DownloadImage(log *slog.Logger, storage ImageSqlSaver, imgStorage img_storage.ImageStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.DownloadImage"

		id := r.URL.Query().Get(idQueryParameter)
		if id == "" {
			log.Error("id parameter is empty", "op", op)
			http.Error(w, "Id parameter is empty", http.StatusBadRequest)
			return
		}
		intID, err := strconv.Atoi(id)
		if err != nil {
			log.Error("id parameter is not number type", "op", op, "err", err)
			http.Error(w, "incorrect id parameter", http.StatusBadRequest)
			return
		}

		// TODO: to call sql storage to get metadata
		// Create check status of gotten image
		metadata, err := storage.DownloadImage(intID)
		if err != nil {
			log.Error("getting data error", "op", op, "err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		// if status is pending or processing
		if metadata.Status != "modified" {
			resp := ImageActionResponse{
				Status:  metadata.Status,
				Message: "Server is handling an image",
				ImageID: intID,
				Action:  metadata.Action,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusProcessing)
			_ = json.NewEncoder(w).Encode(resp)
		}
		// TODO: to call image storage
		image, err := img_storage.GetUpdatedImage(metadata.OriginalPath)
		if err != nil {
			log.Error("Get updated image error", "op", op, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var preparedRespMessage string
		switch metadata.Action {
		case "resize":
			preparedRespMessage = "Imaged was resized"
		case "miniature":
			preparedRespMessage = "Imaged was miniatured"
		case "watermark":
			preparedRespMessage = "Watermark was added to image"
		}

		respWithImage := ImageIncludedResponse{
			Status:  http.StatusText(http.StatusOK),
			Image:   image,
			Message: preparedRespMessage,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(respWithImage)
	}
}

func DeleteImage(log *slog.Logger, storage ImageSqlSaver) func(http.ResponseWriter, *http.Request) {
	return func(http.ResponseWriter, *http.Request) {
		const op = "sqlite.DeleteImage"
	}
}
