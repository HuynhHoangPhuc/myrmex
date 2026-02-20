package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load reads config from file and environment variables.
// Env vars override file values. Prefix is uppercased app name.
func Load(name, path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.AddConfigPath(path)
	v.AddConfigPath(".")
	v.SetEnvPrefix(strings.ToUpper(name))
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	return v, nil
}
