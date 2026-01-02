package handlers

import (
	"fmt"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/models"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
)

type ImageSqlSaver interface {
	UploadImage() error
	DownloadImage() error
	DeleteImage() error
}

// TODO: update
func UploadImage(log *slog.Logger, storage ImageSqlSaver, imgStorage img_storage.ImageStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.UploadImage"

		file, handler, err := r.FormFile("image")
		if err != nil {
			log.Error("error getting file","op", op, "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		// check extension
		extension := filepath.Ext(handler.Filename)

		allowedExtensions := map[string]bool{
			".jpg": true,
			".png": true,
			".gif":  true,
		}

		if !allowedExtensions[extension] {
			log.Warn("%s; file extension not allowed", op)
			http.Error(w, fmt.Sprintf("invalid file extension"), http.StatusBadRequest)
			return
		}

		// upload file into local storage
		newFilePath := filepath.Join(imgStorage.ImgStoragePath, filepath.Base(handler.Filename))
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
			OriginalFilename:
		}

		//TODO: to add kafka producer calling

		fmt.Fprintf(w, "File %s downloaded seccessfuly", handler.Filename)
	}
}

func DownloadImage(log *slog.Logger, storage ImageSqlSaver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "sqlite.DownloadImage"
	}
}

func DeleteImage(log *slog.Logger, storage ImageSqlSaver) func(http.ResponseWriter, *http.Request) {
	return func(http.ResponseWriter, *http.Request) {
		const op = "sqlite.DeleteImage"
	}
}
