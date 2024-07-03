package xconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

type loader struct {
	Env          bool
	Secret       bool
	Path         string
	AwsSsm       bool
	AwsSsmPath   string
	AwsSsmParams map[string]string
}

// load config to struct pointer
// prefixes are parents names path, for recursive calling
func (l *loader) load(dst interface{}, prefixes ...string) error {
	configValue := reflect.Indirect(reflect.ValueOf(dst))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid dst, it should be a struct pointer")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var fieldStruct = configType.Field(i)
		var field = configValue.Field(i)
		var value string
		var source string // for debug

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		// check default value first
		defaultValue := fieldStruct.Tag.Get("default")
		if defaultValue != "" {
			value = defaultValue
			source = "default"
		}
		// check aws ssm
		if l.AwsSsm && l.AwsSsmParams != nil {
			ssmName := fieldStruct.Tag.Get("ssm")
			if ssmName == "" {
				ssmName = strings.ToUpper(strings.Join(append(prefixes, strcase.ToSnake(fieldStruct.Name)), "_"))
			}
			if v, ok := l.AwsSsmParams[ssmName]; ok {
				value = v
				source = "aws_ssm"
			}
		}
		// check shell env
		if l.Env && reflect.Indirect(field).Kind() != reflect.Struct {
			envName := fieldStruct.Tag.Get("env")
			if envName == "" {
				envName = strings.ToUpper(strings.Join(append(prefixes, strcase.ToSnake(fieldStruct.Name)), "_"))
			}
			envValue := os.Getenv(envName)
			if envValue != "" {
				value = envValue
				source = "env"
			}
		}
		// check secret
		if l.Secret {
			secretName := fieldStruct.Tag.Get("secret")
			if secretName == "" {
				secretName = strings.Join(append(prefixes, strcase.ToSnake(fieldStruct.Name)), "_")
			}
			secretValue := l.getSecret(secretName)
			if secretValue != "" {
				value = secretValue
				source = "secret"
			}
		}
		// load value to field
		isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface())
		if isBlank && value != "" {
			slog.Debug("Loading configuration", "field", fieldStruct.Name, "source", source)
			switch reflect.Indirect(field).Kind() {
			case reflect.Bool:
				switch strings.ToLower(value) {
				case "", "0", "f", "false":
					field.Set(reflect.ValueOf(false))
				default:
					field.Set(reflect.ValueOf(true))
				}
			case reflect.String:
				field.Set(reflect.ValueOf(value))
			default:
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			}
		}

		// return error if it is required but blank
		if isBlank && value == "" && fieldStruct.Tag.Get("required") == "true" {
			return errors.New(fieldStruct.Name + " is required")
		}

		// recursive struct and slice
		for field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			// continue
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := l.load(field.Addr().Interface(), fieldNamePath(prefixes, &fieldStruct)...); err != nil {
				return err
			}
		case reflect.Slice:
			if arrLen := field.Len(); arrLen > 0 {
				for i := 0; i < arrLen; i++ {
					if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
						if err := l.load(field.Index(i).Addr().Interface(), append(fieldNamePath(prefixes, &fieldStruct), fmt.Sprint(i))...); err != nil {
							return err
						}
					}
				}
			} else {
				newVal := reflect.New(field.Type().Elem()).Elem()
				if newVal.Kind() == reflect.Struct {
					idx := 0
					for {
						newVal = reflect.New(field.Type().Elem()).Elem()
						if err := l.load(newVal.Addr().Interface(), append(fieldNamePath(prefixes, &fieldStruct), fmt.Sprint(idx))...); err != nil {
							return err
						} else if reflect.DeepEqual(newVal.Interface(), reflect.New(field.Type().Elem()).Elem().Interface()) {
							break
						} else {
							idx++
							field.Set(reflect.Append(field, newVal))
						}
					}
				}
			}
		default:
			// do nothing here, just for linter check
		}
	}

	return nil
}

func (l *loader) getSecret(name string) string {
	data, err := os.ReadFile(path.Join(l.Path, name))
	if os.IsNotExist(err) {
		return ""
	} else if err != nil {
		slog.Error("read secret file error", "error", err)
		return ""
	}
	return strings.TrimSpace(string(data))
}

func fieldNamePath(prefixes []string, fieldStruct *reflect.StructField) []string {
	if fieldStruct.Anonymous && fieldStruct.Tag.Get("anonymous") == "true" {
		return prefixes
	}
	return append(prefixes, strcase.ToSnake(fieldStruct.Name))
}
