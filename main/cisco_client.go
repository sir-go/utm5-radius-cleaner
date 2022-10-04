package main

import (
	"fmt"
	"time"

	expect "github.com/google/goexpect"
)

func CiscoDropSessions(usernames []string) {
	LOG.Printf("cisco: drop %d sessions", len(usernames))
	gexp, _, err := expect.Spawn(fmt.Sprintf("telnet %s", CFG.Cisco.Address), -1)
	eh(err)
	defer func() { _ = gexp.Close() }()

	commands := []expect.Batcher{
		&expect.BExp{R: "Username:"},
		&expect.BSnd{S: CFG.Cisco.Username + "\n"},
		&expect.BExp{R: "Password:"},
		&expect.BSnd{S: CFG.Cisco.Password + "\n"},
	}

	for _, username := range usernames {
		commands = append(commands,
			&expect.BExp{R: "ISG-Router>"},
			&expect.BSnd{S: fmt.Sprintf("clear sss session user %s\n", username)})
	}

	commands = append(commands, &expect.BExp{R: "ISG-Router>"}, &expect.BSnd{S: "logout\n"})

	timeout := time.Duration(
		int64(
			len(commands))*CFG.Cisco.TelnetCmdTimeout.Duration.Milliseconds()) * time.Millisecond
	LOG.Println("cisco: do commands with timeout", timeout)
	if resps, err := gexp.ExpectBatch(commands, timeout); err != nil {
		for _, r := range resps {
			LOG.Println(r)
		}
		eh(err)
	}
	LOG.Println("cisco: done")
}
