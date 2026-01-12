package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Storage        StorageParameters `yaml:"storage"`
	Brokers        []string          `yaml:"brokers" env-required:"true"`
	ImgStoragePath ImageStoragePath  `yaml:"img_storage"`
}

type StorageParameters struct {
	StoragePath string `yaml:"path" env:"STORAGE_PATH" env-required:"true"`
}

type ImageStoragePath struct {
	Path string `yaml:"path" env:"IMG_STORAGE_PATH" env-required:"true"`
}

func MustLoad(pathConfig string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(pathConfig, &cfg)
	if err != nil {
		panic(fmt.Errorf("to set config error; %w", err))
	}

	return &cfg
}
