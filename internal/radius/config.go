package radius

import (
	"time"
)

type Config struct {
	Address  string        `toml:"address"`
	Username string        `toml:"username"`
	Password string        `toml:"password"`
	Timeout  time.Duration `toml:"timeout"`
}
