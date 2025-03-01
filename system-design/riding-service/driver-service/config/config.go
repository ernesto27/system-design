package config

import (
	"github.com/ernesto/riding-service/shared/config"
)

func Load() (*config.Config, error) {
	return config.Load()
}
