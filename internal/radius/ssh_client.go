package radius

import (
	expect "github.com/google/goexpect"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type (
	Client struct {
		cfg Config
	}
)

func NewClient(cfg Config) *Client {
	return &Client{cfg}
}

func (c *Client) Exec(cmd string) error {
	sshClt, err := ssh.Dial(
		"tcp", c.cfg.Address+":22",
		&ssh.ClientConfig{
			User:            c.cfg.Username,
			Auth:            []ssh.AuthMethod{ssh.Password(c.cfg.Password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
	if err != nil {
		return err
	}
	defer func() { _ = sshClt.Close() }()

	exp, _, err := expect.SpawnSSH(sshClt, c.cfg.Timeout)
	if err != nil {
		return err
	}
	defer func() { _ = exp.Close() }()

	if responses, err := exp.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: ":~#"},
		&expect.BSnd{S: cmd + "\n"},
		&expect.BExp{R: ":~#"},
	}, c.cfg.Timeout); err != nil {
		msg := "\n"
		for _, r := range responses {
			msg += r.Output
		}
		return errors.Wrap(err, msg)
	}
	return nil
}
