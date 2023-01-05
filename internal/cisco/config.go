package cisco

import (
	"time"
)

type Config struct {
	Address         string        `toml:"address"`
	Username        string        `toml:"username"`
	Password        string        `toml:"password"`
	CommandTimeout  time.Duration `toml:"command_timeout"`
	DropSessionsCmd string        `toml:"drop_sessions_cmd"`
}
