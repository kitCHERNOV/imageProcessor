package decoding

import (
	"encoding/base64"
	"encoding/json"
	"imageProcessor/internal/handlers"
	"log/slog"
	"net/http"
)

func JSONImageDecoder(r *http.Request, log *slog.Logger) []byte {
	const op = "handlers.decoding.JSONImageDecoder"
	var req handlers.ImageActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request", "op", op, "err", err)
		//http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil
	}

	if err := req.RequestValidate(); err != nil {
		log.Warn("validation failed", "op", op, "err", err)
		//http.Error(w, err.Error(), http.StatusBadRequest)
	}

	imageData, err := base64.StdEncoding.DecodeString(req.Image)
	if err != nil {
		log.Warn("Invalid image format", "op", op, "err", err)
		//http.Error(w, "Invalid image format", http.StatusBadRequest)
		return nil
	}

	return imageData
}

func FormImageDecoding() {}
