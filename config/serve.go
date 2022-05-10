package config

import "time"

type HttpConfig struct {
	Addr           string
	UpdateInterval time.Duration
	ChainConfig    ChainConfig
}

func GetDefaultHttpConfig() HttpConfig {
	return HttpConfig{
		Addr:           "127.0.0.1:8080",
		UpdateInterval: 15 * time.Minute,
		ChainConfig:    GetDefaultChainConfig(),
	}
}
