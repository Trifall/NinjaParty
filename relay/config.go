package main

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	NinjabrainURL  string `mapstructure:"NINJABRAIN_URL"`
	ServerAPIURL   string `mapstructure:"SERVER_API_URL"`
	ServerWSURL    string `mapstructure:"SERVER_WS_URL"`
	LocalUIAddr    string `mapstructure:"LOCAL_UI_ADDR"`
	ObserverURL    string `mapstructure:"OBSERVER_URL"`
	TLSSkipVerify  bool   `mapstructure:"TLS_SKIP_VERIFY"`
	HeartbeatIntvl int    `mapstructure:"HEARTBEAT_INTERVAL"`
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	viper.SetDefault("NINJABRAIN_URL", "http://127.0.0.1:52533")
	viper.SetDefault("SERVER_API_URL", "http://np-server.trifall.com/api/v1")
	viper.SetDefault("SERVER_WS_URL", "wss://np-server.trifall.com/ws")
	viper.SetDefault("LOCAL_UI_ADDR", "127.0.0.1:52532")
	viper.SetDefault("OBSERVER_URL", "https://np.trifall.com/room/")
	viper.SetDefault("TLS_SKIP_VERIFY", false)
	viper.SetDefault("HEARTBEAT_INTERVAL", 10)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
