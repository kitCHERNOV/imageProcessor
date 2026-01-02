package main

import (
	"github.com/joho/godotenv"
	"imageProcessor/internal/config"
	"imageProcessor/internal/handlers"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const configPath = "CONFIG_PATH"

func main() {
	// config creator
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	// gotten config
	cfg := config.MustLoad(os.Getenv(configPath))

	// logger init
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// storage init
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	// image storage
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic("image storage creating error")
	}

	imgStorage := img_storage.ImageStorage{
		ImgStoragePath: uploadDir,
	}

	// TODO:
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/upload", handlers.UploadImage(logger, storage))
	router.Group(func(r chi.Router) {
		r.Get("/image/{id}", handlers.DownloadImage(logger, storage))
		r.Delete("/image/{id}", handlers.DeleteImage(logger, storage))
	})

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}
