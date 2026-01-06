package img_storage

import (
	"fmt"
	"os"
)

type ImageStorage struct {
	ImgStoragePath string
}

func (ims *ImageStorage) ToUpdateImage() {

}

func GetUpdatedImage(filePath string) ([]byte, error) {
	const op = "img-storage.GetUpdatedImage"

	image, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%s,%w", op, err)
	}

	return image, nil
}
