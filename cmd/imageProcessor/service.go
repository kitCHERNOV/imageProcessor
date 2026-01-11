package main

import (
	"errors"
	"fmt"
	"imageProcessor/internal/config"
	"imageProcessor/internal/handlers"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/kafka"
	"imageProcessor/internal/storage/sqlite"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const configPath = "CONFIG_PATH"

// some broker topics
const (
	imgUploadTopic  = "image-upload"
	consumerGroupId = "image-processor"
)

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

	// kafka manager init
	manager, err := kafka.NewKafkaManager(cfg.Brokers)
	if err != nil {
		panic(err)
	}
	defer manager.Close()

	// kafka topics init
	log.Println("Initializing kafka topics...")
	topics := map[string]sarama.TopicDetail{
		imgUploadTopic: {
			NumPartitions:     3,
			ReplicationFactor: 1,
		},
	}
	err = manager.InitTopics(topics)
	if err != nil {
		panic(fmt.Errorf("creating topics failed; error: %w", err))
	}
	// kafka producer init
	producer, err := kafka.NewProducer(cfg.Brokers)
	if err != nil {
		panic(fmt.Errorf("kafka producer does not create; err: %w", err))
	}
	defer producer.Close()

	// kafka consumer init
	//consumer, err := kafka.NewConsumer(cfg.Brokers, consumerGroupId)
	//if err != nil {
	//	panic(fmt.Errorf("kafka consumer does not create; err: %w", err))
	//}
	//defer consumer.Close()

	//uploadDir := "./uploads"
	if err := os.MkdirAll(cfg.ImgStoragePath, os.ModePerm); err != nil {
		panic("image storage creating error")
	}
	// image storage
	imgStorage := img_storage.ImageStorage{
		ImgStoragePath: cfg.ImgStoragePath,
	}

	// TODO:
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/upload", handlers.UploadImage(logger, storage, imgStorage, producer))
	router.Group(func(r chi.Router) {
		r.Get("/image/{id}", handlers.DownloadImage(logger, storage))
		r.Delete("/image/{id}", handlers.DeleteImage(logger, storage))
	})

	doneChannel := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kafka.Consumer(logger, cfg.Brokers, imgUploadTopic, doneChannel, storage); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := http.ListenAndServe(":8080", router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	// Stop server
	close(doneChannel)
	wg.Wait()
}
