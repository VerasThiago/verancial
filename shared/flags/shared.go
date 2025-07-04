package flags

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	DEPLOY_ENV string = "VERANCIAL_DEPLOY_ENV"

	DEPLOY_LOCAL      string = "local"
	DEPLOY_DOCKER     string = "docker"
	DEPLOY_PRODUCTION string = "production"

	ENV_FILE_PATH string = ".env"
	ENV_FILE_TYPE string = "env"

	SHARED_PACKAGE_NAME = "shared"
	SHARED_PACKAGE_PATH = "../shared"
)

var AVAILABLE_ENV_VARS_MAP = map[string]bool{
	"production": true,
	"local":      true,
	"docker":     true,
}

type SharedFlags struct {
	Deploy           string
	AppHost          string `mapstructure:"APP_HOST"`
	DatabaseURL      string `mapstructure:"DATABASE_URL"`
	DatabaseHost     string `mapstructure:"DB_HOST"`
	DatabasePort     string `mapstructure:"DB_PORT"`
	DatabaseUser     string `mapstructure:"DB_USER"`
	DatabasePassword string `mapstructure:"DB_PASSWORD"`
	DatabaseName     string `mapstructure:"DB_NAME"`
	DatabaseSSLMode  string `mapstructure:"DB_SSLMODE"`
	DatabaseTimeZone string `mapstructure:"DB_TIMEZONE"`
	QueueHost        string `mapstructure:"QUEUE_HOST"`
	QueuePort        string `mapstructure:"QUEUE_PORT"`
	JwtKey           string `mapstructure:"JWT_KEY"`
	JwtKeyEmail      string `mapstructure:"JWT_KEY_EMAIL"`
	DPWHost          string `mapstructure:"DPW_HOST"`
	DPWPort          string `mapstructure:"DPW_PORT"`
	AIWHost          string `mapstructure:"AIW_HOST"`
	AIWPort          string `mapstructure:"AIW_PORT"`
}

type EnvFileConfig struct {
	Path string
	Name string
	Type string
}

func GetDeployEnv() string {
	env := os.Getenv(DEPLOY_ENV)
	if _, ok := AVAILABLE_ENV_VARS_MAP[string(env)]; !ok {
		panic("invalid VERANCIAL_DEPLOY_ENV env variable i.e. [production, local, docker]")
	}
	return env
}

func GetFileEnvConfigFromDeployEnv(serviceName string) *EnvFileConfig {
	deployEnv := GetDeployEnv()

	var filePath string = ENV_FILE_PATH
	fileName := fmt.Sprintf("%v.%v.env", serviceName, deployEnv)

	if serviceName == SHARED_PACKAGE_NAME {
		filePath = fmt.Sprintf("%+v/%+v", SHARED_PACKAGE_PATH, ENV_FILE_PATH)
	}

	return &EnvFileConfig{
		Path: filePath,
		Name: fileName,
		Type: ENV_FILE_TYPE,
	}
}

func (f *SharedFlags) InitFromViper(config *EnvFileConfig) (*SharedFlags, error) {
	viper := viper.New()
	viper.AddConfigPath(config.Path)
	viper.SetConfigName(config.Name)
	viper.SetConfigType(config.Type)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var flags SharedFlags
	if err := viper.Unmarshal(&flags); err != nil {
		return nil, err
	}

	flags.Deploy = GetDeployEnv()

	return &flags, nil
}
