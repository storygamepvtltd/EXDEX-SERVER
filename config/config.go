package config

import (
	"log"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("app")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("error on parsing configuration file: %s", err)
	}
}
