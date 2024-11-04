package config

import "github.com/spf13/viper"

type Config struct {
	ServerAddress string        `json:"server_address"`
	DatabaseURL   string        `json:"database_url"`
	StoragePath   string        `json:"storage_path"`
	JWTSecret     string        `json:"jwt_secret"`
	Storage       StorageConfig `mapstructure:"storage"`
}

type StorageConfig struct {
	Type  string      `mapstructure:"type"`
	Minio MinioConfig `mapstructure:"minio"`
	Local LocalConfig `mapstructure:"local"`
}

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type LocalConfig struct {
	Path string `mapstructure:"path"`
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
