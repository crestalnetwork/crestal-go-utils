package xdb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/samber/oops"
)

var awsConfig *aws.Config

type asmRes struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
}

func loadFromAwsSecretsManager(cfg *Config) error {
	ctx := context.Background()
	// aws client
	if awsConfig == nil {
		// try to get default aws config
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("load default aws config failed: %w", err)
		}
		awsConfig = &cfg
	}
	client := secretsmanager.NewFromConfig(*awsConfig)
	res, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.SecretsManagerPath),
	})
	if err != nil {
		return oops.Wrapf(err, "get secret value from aws failed")
	}
	// parse
	var asm asmRes
	err = json.Unmarshal([]byte(*res.SecretString), &asm)
	if err != nil {
		return oops.Wrapf(err, "unmarshal secret string failed")
	}
	// set
	cfg.Host = asm.Host
	cfg.Port = fmt.Sprintf("%d", asm.Port)
	cfg.Username = asm.Username
	cfg.Password = asm.Password
	if cfg.Name == "" {
		// if user set the Name field, it will override the DB Name field in SecretManager value
		cfg.Name = asm.DBName
	}
	return nil
}
