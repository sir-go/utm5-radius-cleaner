package main

import (
	"flag"

	zlog "github.com/rs/zerolog/log"

	"radCleaner/internal/cisco"
	"radCleaner/internal/radius"
	"radCleaner/internal/utm"
)

const DefaultConfFile = "config.toml"

func eh(err error) {
	if err != nil {
		zlog.Err(err).Msg("")
		panic(err)
	}
}

func main() {
	// parse -c parameter
	fCfgPath := flag.String("c", DefaultConfFile, "path to conf file")
	flag.Parse()

	// load the config file
	cfg, err := LoadConfig(*fCfgPath)
	eh(err)

	utmClient := utm.NewClient(cfg.UTM)

	// get all active UTM Radius sessions
	sessions, err := utmClient.GetActiveRadiusSessions()
	eh(err)

	// usernames of zero-address sessions for drop on the Cisco
	usernames := make([]string, 0)

	// expired sessions for drop on the UTM
	utmExpiredSessions := make([]utm.RadiusSession, 0)

	for _, s := range sessions {

		if s.FramedIp4 == "0.0.0.0" {
			usernames = append(usernames, s.UserName)
		}

		// if session is expired - try to drop it on the UTM
		if expired, since := s.IsExpired(cfg.SessionLifeTime); expired {
			zlog.Debug().Dur("since", since).Str("username", s.UserName).Msg("expired")
			if err = utmClient.DropRadiusSession(s); err != nil {
				zlog.Err(err).Msg("drop utm session")
			}

			// store it for retry if failed
			utmExpiredSessions = append(utmExpiredSessions, s)
		}
	}

	// drop zero-address sessions on the Cisco
	ciscoClient := cisco.NewClient(cfg.Cisco)
	eh(ciscoClient.DropSessions(usernames))

	// if no expired sessions - exit
	if len(utmExpiredSessions) < 1 {
		zlog.Debug().Msg("utm: nothing was dropped, exit")
		return
	}

	// get new list of active sessions
	newSessionsList, err := utmClient.GetActiveRadiusSessions()
	eh(err)

	// check which expired sessions still active
	stillSessions := utm.SessionsIntersect(newSessionsList, utmExpiredSessions)
	if len(stillSessions) < 1 {
		zlog.Debug().Msg("utm: no one sessions, exit")
		return
	}

	// if there are expired sessions - restart the RADIUS server via SSH
	sshClient := radius.NewClient(cfg.SSH)
	eh(sshClient.Exec(cfg.RadiusRestartCmd))

	// retry to drop expired session after RADIUS server restart
	for _, s := range stillSessions {
		if err = utmClient.DropRadiusSession(s); err != nil {
			zlog.Err(err).Msg("drop utm session")
		}
	}
}
