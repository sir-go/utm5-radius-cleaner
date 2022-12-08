package cisco

import (
	"fmt"
	"time"

	expect "github.com/google/goexpect"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

type (
	Config struct {
		Address         string        `toml:"address"`
		Username        string        `toml:"username"`
		Password        string        `toml:"password"`
		CommandTimeout  time.Duration `toml:"command_timeout"`
		DropSessionsCmd string        `toml:"drop_sessions_cmd"`
	}
	Client struct {
		Config
	}
)

func NewClient(cfg Config) *Client {
	return &Client{cfg}
}

func (c *Client) DropSessions(usernames []string) error {
	if len(usernames) < 1 {
		zlog.Debug().Msg("cisco: nothing to drop")
		return nil
	}
	gExpect, _, err := expect.Spawn("telnet "+c.Address, -1)
	if err != nil {
		return err
	}
	defer func() { _ = gExpect.Close() }()

	commands := []expect.Batcher{
		&expect.BExp{R: "Username:"},
		&expect.BSnd{S: c.Username + "\n"},
		&expect.BExp{R: "Password:"},
		&expect.BSnd{S: c.Password + "\n"},
	}

	for _, username := range usernames {
		commands = append(commands,
			&expect.BExp{R: "ISG-Router>"},
			&expect.BSnd{S: fmt.Sprintf("%s %s\n", c.DropSessionsCmd, username)})
	}

	commands = append(commands, &expect.BExp{R: "ISG-Router>"}, &expect.BSnd{S: "logout\n"})
	timeout := time.Duration(int64(len(commands))) * c.CommandTimeout

	if responses, err := gExpect.ExpectBatch(commands, timeout); err != nil {
		msg := "\n"
		for _, r := range responses {
			msg += r.Output
		}
		return errors.Wrap(err, msg)
	}
	return nil
}
