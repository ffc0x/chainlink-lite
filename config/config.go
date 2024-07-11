package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database    Database    `mapstructure:"database"`
	PriceTicker PriceTicker `mapstructure:"price_ticker"`
	PubSub      PubSub      `mapstructure:"pubsub"`
	LogLevel    int         `mapstructure:"log_level"`
}

type Database struct {
	URL string `mapstructure:"url"`
}

type PriceTicker struct {
	URL  string `mapstructure:"url"`
	Mock bool   `mapstructure:"mock"`
}

type PubSub struct {
	TopicName                string        `mapstructure:"topic"`
	FetchPriceInterval       time.Duration `mapstructure:"fetch_price_interval"`
	MinSignaturesToWrite     int           `mapstructure:"min_signatures_to_write"`
	MinIntervalBetweenWrites time.Duration `mapstructure:"min_interval_between_writes"`
	DiscoverPeersInterval    time.Duration `mapstructure:"discover_peers_interval"`
	Port                     int           `mapstructure:"port"`
}

func LoadConfig() (Config, error) {
	// Set the path to look for the configurations
	viper.AddConfigPath("config")
	viper.AddConfigPath("../../config")
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Automatically read configuration from environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
