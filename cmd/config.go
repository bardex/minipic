package main

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server struct {
		Listen string
	}
	Cache struct {
		Limit     int
		Directory string
	}
}

func NewConfig(configPath string) (Config, error) {
	config := Config{}
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, err
	}
	return config, nil
}
