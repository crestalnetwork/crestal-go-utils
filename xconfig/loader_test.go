package xconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	AppName string `default:"default_app"` // testing snake env APP_NAME
	DB      struct {
		Name     string `default:"default_name"`             // testing default
		User     string `default:"default_user"`             // testing env
		Password string `default:"default_pwd"`              // testing secret
		Port     int    `default:"3306" env:"MYSQL_DB_PORT"` // testing int and custom env name
	}
}

func TestLoadingConfig(t *testing.T) {
	var err error
	var l = loader{
		Debug:  true,
		Env:    true,
		Secret: true,
		Path:   "mock",
	}

	// mock shell env variables
	err = os.Setenv("APP_NAME", "env_app")
	if err != nil {
		t.Error(err)
		return
	}
	err = os.Setenv("DB_USER", "env_user")
	if err != nil {
		t.Error(err)
		return
	}
	err = os.Setenv("DB_PASSWORD", "env_pwd")
	if err != nil {
		t.Error(err)
		return
	}
	err = os.Setenv("MYSQL_DB_PORT", "3307")
	if err != nil {
		t.Error(err)
		return
	}

	// load
	cfg := new(testConfig)
	err = l.load(cfg)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, "env_app", cfg.AppName)
	assert.Equal(t, "default_name", cfg.DB.Name)
	assert.Equal(t, "env_user", cfg.DB.User)
	assert.Equal(t, "secret_pwd", cfg.DB.Password)
	assert.Equal(t, 3307, cfg.DB.Port)
}
