package main

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ListenAddr string        `mapstructure:"LISTEN_ADDR"`
	UIDomain   string        `mapstructure:"UI_DOMAIN"`
	RoomTTL    time.Duration `mapstructure:"ROOM_TTL"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("LISTEN_ADDR", ":8080")
	viper.SetDefault("UI_DOMAIN", "http://localhost:5173")
	viper.SetDefault("ROOM_TTL", 1*time.Hour) // 1 hour inactivity TTL

	viper.AutomaticEnv()
	// allow UI_DOMAIN=http://foo,http://bar
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
