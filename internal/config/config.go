package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	StorageParameters `yaml:"storage"`
	Brokers           []string `yaml:"brokers" env-required:"true"`
	ImgStoragePath    string   `yaml:"img_storage"`
}

type StorageParameters struct {
	StoragePath string `yaml:"path" env:"STORAGE_PATH" env-require:"true"`
}

type ImageStoragePath struct {
	Path string `yaml:"path" env:"IMG_STORGE_PATH" env-required:"true"`
}

func MustLoad(pathConfing string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(pathConfing, &cfg)
	if err != nil {
		panic(fmt.Errorf("to set config error; %w", err))
	}

	return &cfg
}
