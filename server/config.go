package main

import (
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	ListenAddr  string        `mapstructure:"LISTEN_ADDR"`
	ObserverURL string        `mapstructure:"OBSERVER_URL"`
	RoomTTL     time.Duration `mapstructure:"ROOM_TTL"`
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	viper.SetDefault("LISTEN_ADDR", ":8080")
	viper.SetDefault("OBSERVER_URL", "https://np.trifall.com/")
	viper.SetDefault("ROOM_TTL", 2*time.Hour) // 2 hour inactivity TTL

	viper.AutomaticEnv()
	// allow UI_DOMAIN=http://foo,http://bar
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
