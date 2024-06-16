package xconfig

// Load config to `dst` struct pointer from shell env variables and docker secrets.
func Load(dst interface{}) error {
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
