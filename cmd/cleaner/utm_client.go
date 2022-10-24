package main

import (
	"fmt"
	"time"

	"github.com/KeisukeYamashita/jsonrpc"
	expect "github.com/google/goexpect"
	"golang.org/x/crypto/ssh"
)

type (
	rActiveRadiusSession struct {
		FramedIp4      string `json:"framed_ip4"`
		AcctSessionId  string `json:"traf_acct_session_id"`
		Id             int    `json:"traf_id"`
		LastUpdateDate int64  `json:"traf_last_update_date"`
		NasIp          string `json:"traf_nas_ip"`
		RecvDate       int64  `json:"traf_recv_date"`
		UserName       string `json:"traf_user_name"`
	}

	rActiveRadiusSessions struct {
		Sessions []rActiveRadiusSession `json:"traffic_sessions"`
	}

	UtmApi struct {
		BillingPrefix string
	}

	UtmArgs map[string]interface{}
)

func sessionsIntersect(n, o []rActiveRadiusSession) (still []rActiveRadiusSession) {
	still = make([]rActiveRadiusSession, 0)
	for _, oldSess := range o {
		for _, newSess := range n {
			if oldSess.AcctSessionId == newSess.AcctSessionId && oldSess.UserName == newSess.UserName {
				still = append(still, oldSess)
			}
		}
	}
	return
}

func (s *rActiveRadiusSession) isExpired() (expired bool, since time.Duration) {
	since = time.Since(time.Unix(s.LastUpdateDate, 0))
	expired = since > CFG.Billing.SessionLifeTime.Duration
	return
}

func (u *UtmApi) call(method string, args UtmArgs, target interface{}) (err error) {
	LOG.Println(method, args)
	var res *jsonrpc.RPCResponse
	client := jsonrpc.NewRPCClient(CFG.Billing.ApiURL)
	client.SetBasicAuth(CFG.Billing.Username, CFG.Billing.Password)

	if res, err = client.CallNamed(u.BillingPrefix+"."+method, args); err != nil {
		return
	}

	if res.Error != nil {
		err = fmt.Errorf("urfa api error: %d : %s : %v",
			res.Error.Code, res.Error.Message, res.Error.Data)
		return
	}

	if target != nil {
		return res.GetObject(target)
	}
	return
}

func (u *UtmApi) GetActiveRadiusSessions() ([]rActiveRadiusSession, error) {
	o := new(rActiveRadiusSessions)
	if err := u.call("rpcf_radius_get_active_sessions", UtmArgs{}, o); err != nil {
		return nil, err
	}
	return o.Sessions, nil
}

func (u *UtmApi) DropRadiusSession(sessionId string, nasIp string) error {
	return u.call("rpcf_radius_drop_session", UtmArgs{
		"acct_session_id": sessionId, "nas_ip": nasIp}, nil)
}

func (u *UtmApi) RestartRadius() {
	LOG.Println("restart RADIUS")

	sshClt, err := ssh.Dial(
		"tcp", CFG.Billing.Hosts[u.BillingPrefix]+":22",
		&ssh.ClientConfig{
			User:            CFG.Billing.SSHUsername,
			Auth:            []ssh.AuthMethod{ssh.Password(CFG.Billing.SSHPassword)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
	eh(err)
	defer func() { _ = sshClt.Close() }()

	exp, _, err := expect.SpawnSSH(sshClt, CFG.Billing.SSHTimeout.Duration)
	eh(err)
	defer func() { _ = exp.Close() }()

	if resps, err := exp.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: ":~#"},
		&expect.BSnd{S: CFG.Billing.RadiusRestartCmd + "\n"},
		&expect.BExp{R: ":~#"},
	}, CFG.Billing.SSHTimeout.Duration); err != nil {
		for _, r := range resps {
			LOG.Println(r)
		}
		eh(err)
	}
	LOG.Println("RADIUS restarted")
}

func (u *UtmApi) CleanSessions() {
	LOG.SetPrefix(u.BillingPrefix + ": ")
	defer LOG.SetPrefix("")

	LOG.Println("clean sessions")
	sessions, err := u.GetActiveRadiusSessions()
	eh(err)
	LOG.Printf("got %d active session", len(sessions))
	toDropOnCisco := make([]string, 0)
	utmDropped := make([]rActiveRadiusSession, 0)
	for _, s := range sessions {
		if s.FramedIp4 == "0.0.0.0" {
			LOG.Println("0.0.0.0 IPv4", s)
			toDropOnCisco = append(toDropOnCisco, s.UserName)
		}
		if expired, since := s.isExpired(); expired {
			LOG.Println("expired", since, s)
			ehSkip(u.DropRadiusSession(s.AcctSessionId, s.NasIp))
			utmDropped = append(utmDropped, s)
		}
	}
	if len(toDropOnCisco) < 1 {
		return
	}

	CiscoDropSessions(toDropOnCisco)

	if len(utmDropped) < 1 {
		return
	}

	LOG.Println("check utm sessions")
	newSessions, err := u.GetActiveRadiusSessions()
	eh(err)
	LOG.Printf("got %d active session", len(newSessions))
	stillSessions := sessionsIntersect(newSessions, utmDropped)
	if len(stillSessions) < 1 {
		LOG.Println("all sessions successfully dropped")
		return
	}

	u.RestartRadius()

	for _, s := range stillSessions {
		ehSkip(u.DropRadiusSession(s.AcctSessionId, s.NasIp))
	}
}
