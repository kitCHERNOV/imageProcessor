package handlers

import (
	"encoding/json"
	"fmt"
	img_storage "imageProcessor/internal/img-storage"
	"imageProcessor/internal/models"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
)

type mockStorage struct {
}

func (ms *mockStorage) SetMetadata(metadata *models.ImageMetadata) (int, error) {
	if metadata == nil {
		return 0, fmt.Errorf("test error")
	}
	return 1, nil
}

func (ms *mockStorage) DownloadImage(id int) (*models.ImageMetadata, error) {
	return &models.ImageMetadata{
		OriginalFilename: "img1.png",
		OriginalPath:     "./uploads/img1.png",
		MimeType:         "png",
		FileSize:         3 * 1024 * 1024,
		Status:           "pending",
		Action:           "resize",
	}, nil
}

func (ms *mockStorage) DeleteImage(id int) error {
	return nil
}

func TestDownloadImage(t *testing.T) {
	type args struct {
		image  []byte
		action string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				image:  []byte("102323"),
				action: "resize",
			},
		},
	}

	mux := http.NewServeMux()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelError,
	}))

	storage := &mockStorage{}
	imgStorage := img_storage.ImageStorage{ImgStoragePath: "./uploads"}

	mux.HandleFunc("/", DownloadImage(log, storage, imgStorage))

	go func() {
		if err := http.ListenAndServe("localhost:8080", mux); err != nil {
			panic(err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.args)
			if err != nil {
				t.Error(err)
			}
			// TODO: to fix body parameter in test post request
			resp, err := http.Post("/", "application/json", body)
			if err != nil {
				t.Error("post request failed")
			}

			var respStruct ImageIncludedResponse
			//if respBody.Status
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			json.Unmarshal(body, &respStruct)
		})
	}
}
