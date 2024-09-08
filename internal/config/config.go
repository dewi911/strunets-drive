package config

import "github.com/spf13/viper"

type Config struct {
	ServerAddress string `json:"server_address"`
	DatabaseURL   string `json:"database_url"`
	StoragePath   string `json:"storage_path"`
	JWTSecret     string `json:"jwt_secret"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
