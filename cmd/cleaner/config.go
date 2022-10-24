package main

import (
	"flag"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type (
	Duration struct {
		time.Duration
	}

	CfgBilling struct {
		ApiURL           string            `toml:"api_url"`
		Username         string            `toml:"username"`
		Password         string            `toml:"password"`
		SessionLifeTime  *Duration         `toml:"session_life_time"`
		SSHUsername      string            `toml:"ssh_username"`
		SSHPassword      string            `toml:"ssh_password"`
		SSHTimeout       *Duration         `toml:"ssh_timeout"`
		RadiusRestartCmd string            `toml:"radius_restart_cmd"`
		Hosts            map[string]string `toml:"hosts"`
	}

	CfgCisco struct {
		Address          string    `toml:"address"`
		Username         string    `toml:"username"`
		Password         string    `toml:"password"`
		TelnetCmdTimeout *Duration `toml:"telnet_cmd_timeout"`
	}

	Config struct {
		Billing CfgBilling `toml:"billing"`
		Cisco   CfgCisco   `toml:"cisco"`
		Path    string
	}
)

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func ConfigInit() *Config {
	fCfgPath := flag.String("c", DefaultConfFile, "path to conf file")
	flag.Parse()

	conf := new(Config)
	file, err := os.Open(*fCfgPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if file == nil {
			return
		}
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err = toml.DecodeFile(*fCfgPath, &conf); err != nil {
		panic(err)
	}
	conf.Path = *fCfgPath
	return conf
}
