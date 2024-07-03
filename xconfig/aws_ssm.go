package xconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// AwsSsmParamStorePath the environment variable name for AWS SSM Param Store
const AwsSsmParamStorePath = "AWS_SSM_PARAM_STORE_PATH"

var awsConfig *aws.Config

type AwsSsmParamStore struct {
	Path string
}

// loadAwsSsmParamStore preload AWS SSM Param Store to memory
func (l *loader) loadAwsSsmParamStore() error {
	ctx := context.Background()
	if awsConfig == nil {
		// try to get default aws config
		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile("default"))
		if err != nil {
			return fmt.Errorf("load default aws config failed: %w", err)
		}
		awsConfig = &cfg
	}
	client := ssm.NewFromConfig(*awsConfig)
	res, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:           aws.String(l.AwsSsmPath),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	if l.AwsSsmParams == nil {
		l.AwsSsmParams = make(map[string]string)
	}
	for _, param := range res.Parameters {
		if param.Name != nil && param.Value != nil {
			name := strings.TrimPrefix(*param.Name, l.AwsSsmPath+"/")
			l.AwsSsmParams[name] = *param.Value
		}
	}

	return nil
}
