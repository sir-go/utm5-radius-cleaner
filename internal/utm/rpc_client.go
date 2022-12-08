package utm

import (
	"fmt"
	"time"

	"github.com/KeisukeYamashita/jsonrpc"
	zlog "github.com/rs/zerolog/log"
)

type (
	RadiusSession struct {
		FramedIp4      string `json:"framed_ip4"`
		AcctSessionId  string `json:"traf_acct_session_id"`
		Id             int    `json:"traf_id"`
		LastUpdateDate int64  `json:"traf_last_update_date"`
		NasIp          string `json:"traf_nas_ip"`
		RecvDate       int64  `json:"traf_recv_date"`
		UserName       string `json:"traf_user_name"`
	}

	RadiusSessions struct {
		Sessions []RadiusSession `json:"traffic_sessions"`
	}

	Config struct {
		Url      string `toml:"api_url"`
		Prefix   string `toml:"prefix"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	}

	Client struct {
		Config
	}

	Args map[string]interface{}
)

func NewClient(cfg Config) *Client {
	return &Client{cfg}
}

func SessionsIntersect(n, o []RadiusSession) (still []RadiusSession) {
	still = make([]RadiusSession, 0)
	for _, oldSess := range o {
		for _, newSess := range n {
			if oldSess.AcctSessionId == newSess.AcctSessionId && oldSess.UserName == newSess.UserName {
				still = append(still, oldSess)
			}
		}
	}
	return
}

func (s *RadiusSession) IsExpired(lifeTime time.Duration) (expired bool, since time.Duration) {
	since = time.Since(time.Unix(s.LastUpdateDate, 0))
	expired = since > lifeTime
	return
}

func (c *Client) call(method string, args Args, target interface{}) (err error) {
	zlog.Debug().Str("method", method).Interface("args", args).Msg("utm call")
	var res *jsonrpc.RPCResponse
	client := jsonrpc.NewRPCClient(c.Url)
	client.SetBasicAuth(c.Username, c.Password)

	if res, err = client.CallNamed(c.Prefix+"."+method, args); err != nil {
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

func (c *Client) GetActiveRadiusSessions() ([]RadiusSession, error) {
	o := new(RadiusSessions)
	if err := c.call("rpcf_radius_get_active_sessions", Args{}, o); err != nil {
		return nil, err
	}
	return o.Sessions, nil
}

func (c *Client) DropRadiusSession(s RadiusSession) error {
	return c.call("rpcf_radius_drop_session", Args{
		"acct_session_id": s.Id, "nas_ip": s.NasIp}, nil)
}
