package cisco

import (
	"fmt"
	"time"

	expect "github.com/google/goexpect"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

type (
	Client struct {
		cfg Config
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
	gExpect, _, err := expect.Spawn("telnet "+c.cfg.Address, -1)
	if err != nil {
		return err
	}
	defer func() { _ = gExpect.Close() }()

	commands := []expect.Batcher{
		&expect.BExp{R: "Username:"},
		&expect.BSnd{S: c.cfg.Username + "\n"},
		&expect.BExp{R: "Password:"},
		&expect.BSnd{S: c.cfg.Password + "\n"},
	}

	for _, username := range usernames {
		commands = append(commands,
			&expect.BExp{R: "ISG-Router>"},
			&expect.BSnd{S: fmt.Sprintf("%s %s\n", c.cfg.DropSessionsCmd, username)})
	}

	commands = append(commands, &expect.BExp{R: "ISG-Router>"}, &expect.BSnd{S: "logout\n"})
	timeout := time.Duration(int64(len(commands))) * c.cfg.CommandTimeout

	if responses, err := gExpect.ExpectBatch(commands, timeout); err != nil {
		msg := "\n"
		for _, r := range responses {
			msg += r.Output
		}
		return errors.Wrap(err, msg)
	}
	return nil
}
