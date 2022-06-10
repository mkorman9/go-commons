package utils

import (
	"github.com/gookit/config/v2"
	"os"
	"strings"
)

func LoadConfigFromEnv() error {
	var envs []string
	for _, env := range os.Environ() {
		s := strings.Split(env, "=")
		envs = append(envs, s[0])
	}

	c := config.NewWithOptions("envs", config.ParseEnv, config.Delimiter(byte('_')))
	c.LoadOSEnv(envs, true)

	return config.LoadData(c.Data())
}
