package xconfig

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// SetAwsConfig if you want load config from aws ssm parameter store, you can set aws config manually.
// If it is not set, it will try to use default aws config, only when AWS_SSM_PARAM_STORE_PATH is set .
func SetAwsConfig(config *aws.Config) {
	awsConfig = config
}

// Load config to `dst` struct pointer from shell env variables and docker secrets.
func Load(dst interface{}) error {
	ssmPath := os.Getenv(AwsSsmParamStorePath)
	if ssmPath != "" {
		err := LoadEnvAndAwsSsm(dst, ssmPath)
		if err != nil {
			return err
		}
		return nil
	}
	err := LoadEnvAndDockerSecret(dst)
	if err != nil {
		return err
	}
	return nil
}

// MustLoad just same as Load(), but it panics when an error occurs.
func MustLoad(dst interface{}) {
	err := Load(dst)
	if err != nil {
		panic(err)
	}
}

// LoadEnvAndSecret load config to `dst` struct pointer from shell env variables and container secrets.
func LoadEnvAndSecret(dst interface{}, secretPath string) error {
	l := loader{
		Env:    true,
		Secret: true,
		Path:   secretPath,
	}
	return l.load(dst)
}

// LoadEnvAndDockerSecret load config to `dst` struct pointer from shell env variables and docker secrets.
func LoadEnvAndDockerSecret(dst interface{}) error {
	return LoadEnvAndSecret(dst, "/run/secrets")
}

// LoadEnv load config to `dst` struct pointer from shell env variables only
func LoadEnv(dst interface{}) error {
	l := loader{
		Env:    true,
		Secret: false,
	}
	return l.load(dst)
}

// LoadEnvAndAwsSsm load config to `dst` struct pointer from shell env variables and aws ssm param store.
func LoadEnvAndAwsSsm(dst interface{}, path string) error {
	l := loader{
		Env:        true,
		AwsSsm:     true,
		AwsSsmPath: path,
	}
	err := l.loadAwsSsmParamStore()
	if err != nil {
		return err
	}
	return l.load(dst)
}
