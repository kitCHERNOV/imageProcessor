package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	StorageParameters `yaml:"storage"`
}

type StorageParameters struct {
	StoragePath string `yaml:"path" env:"STORAGE_PATH" env-require:"true"`
}

func MustLoad(pathConfing string) *Config {
	var cfg Config

	err := cleanenv.ReadConfig(pathConfing, &cfg)
	if err != nil {
		panic(fmt.Errorf("to set config error; %w", err))
	}

	return &cfg
}
