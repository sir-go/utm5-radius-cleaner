package main

import (
	"time"

	"github.com/BurntSushi/toml"

	"radCleaner/internal/cisco"
	"radCleaner/internal/radius"
	"radCleaner/internal/utm"
)

type (
	Config struct {
		SessionLifeTime  time.Duration `toml:"session_life_time"`
		RadiusRestartCmd string        `toml:"radius_restart_cmd"`

		UTM   utm.Config    `toml:"utm"`
		SSH   radius.Config `toml:"ssh"`
		Cisco cisco.Config  `toml:"cisco"`
		Path  string
	}
)

// LoadConfig parses a fileName file to Config structure
func LoadConfig(fileName string) (*Config, error) {
	var conf = new(Config)
	if _, err := toml.DecodeFile(fileName, &conf); err != nil {
		return nil, err
	}
	conf.Path = fileName
	return conf, nil
}
