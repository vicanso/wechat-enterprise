package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// ENV 配置的环境变量
var env = os.Getenv("GO_ENV")
var configPath = os.Getenv("CONFIG")
var viperInitTest = os.Getenv("VIPER_INIT_TEST")

// 初始化配置
func viperInit(path string) error {
	configType := "yml"
	defaultPath := "."
	v := viper.New()
	v.SetConfigName("default")
	v.AddConfigPath(defaultPath)
	v.AddConfigPath(path)
	v.SetConfigType(configType)
	err := v.ReadInConfig()
	if err != nil {
		return err
	}
	configs := v.AllSettings()
	for k, v := range configs {
		viper.SetDefault(k, v)
	}
	if env != "" {
		viper.SetConfigName(env)
		viper.AddConfigPath(defaultPath)
		viper.AddConfigPath(path)
		viper.SetConfigType(configType)
		err := viper.ReadInConfig()
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	if configPath == "" {
		runPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		configPath = runPath + "/configs"
	}
	err := viperInit(configPath)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

// GetENV get the go env
func GetENV() string {
	return env
}

// GetInt viper get int
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetIntDefault get int with default value
func GetIntDefault(key string, defaultValue int) int {
	v := viper.GetInt(key)
	if v != 0 {
		return v
	}
	return defaultValue
}

// GetString viper get string
func GetString(key string) string {
	return viper.GetString(key)
}

// GetStringDefault get string with default value
func GetStringDefault(key, defaultValue string) string {
	v := viper.GetString(key)
	if v != "" {
		return v
	}
	return defaultValue
}

// GetDuration viper get duration
func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// GetDurationDefault get duration with default value
func GetDurationDefault(key string, defaultValue time.Duration) time.Duration {
	v := viper.GetDuration(key)
	if v.Nanoseconds() != 0 {
		return v
	}
	return defaultValue
}

// GetStringSlice viper get string slice
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}
