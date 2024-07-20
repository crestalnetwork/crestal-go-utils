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
	// aws client
	if awsConfig == nil {
		// try to get default aws config
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return fmt.Errorf("load default aws config failed: %w", err)
		}
		awsConfig = &cfg
	}
	client := ssm.NewFromConfig(*awsConfig)

	// init
	if l.AwsSsmParams == nil {
		l.AwsSsmParams = make(map[string]string)
	}

	// first request
	res, err := client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
		Path:           aws.String(l.AwsSsmPath),
		WithDecryption: aws.Bool(true),
	})
	// process then pagination
	for {
		if err != nil {
			return err
		}
		for _, param := range res.Parameters {
			if param.Name != nil && param.Value != nil {
				name := strings.TrimPrefix(*param.Name, l.AwsSsmPath+"/")
				l.AwsSsmParams[name] = *param.Value
			}
		}
		if res.NextToken != nil {
			res, err = client.GetParametersByPath(ctx, &ssm.GetParametersByPathInput{
				Path:           aws.String(l.AwsSsmPath),
				WithDecryption: aws.Bool(true),
				NextToken:      res.NextToken,
			})
		} else {
			break
		}
	}

	return nil
}
