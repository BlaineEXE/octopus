package config

import (
	"log"
	"reflect"

	"github.com/spf13/viper"
)

// Read from the config file
func loadConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./.octopus")
	viper.AddConfigPath("$HOME/.octopus")
	viper.AddConfigPath("/etc/octopus/")
	err := viper.ReadInConfig()
	if err != nil && !isConfigFileNotFoundError(err) {
		log.Fatalf("Error reading config file: %+v", err)
	}
}

func isConfigFileNotFoundError(err error) bool {
	return reflect.TypeOf(err) == reflect.TypeOf(viper.ConfigFileNotFoundError{})
}
