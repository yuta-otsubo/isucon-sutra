package webapp

import (
	"crypto/tls"
	"net/http"
	"time"

	// isucandarはISUCONなどの負荷試験で使える機能を集めたフレームワーク
	"github.com/isucon/isucandar/agent"
	// zapはGoの高性能なロギングライブラリ
	"go.uber.org/zap"
)

type Client struct {
	agent *agent.Agent

	contestantLogger *zap.Logger

	requestModifier []func(*http.Request)
}

type ClientConfig struct {
	TargetBaseURL         string
	DefaultClientTimeout  time.Duration
	ClientIdleConnTimeout time.Duration
	InsecureSkipVerify    bool
	ContestantLogger      *zap.Logger
}

func NewClient(config ClientConfig) (*Client, error) {
	ag, err := agent.NewAgent(
		agent.WithBaseURL(config.TargetBaseURL),
		agent.WithTimeout(config.DefaultClientTimeout),
		agent.WithNoCache(),
		agent.WithCloneTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.InsecureSkipVerify,
			},
			IdleConnTimeout:   config.ClientIdleConnTimeout,
			ForceAttemptHTTP2: true,
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		agent:            ag,
		contestantLogger: config.ContestantLogger,
	}, nil
}

func (c *Client) AddRequestModifier(modifier func(*http.Request)) {
	c.requestModifier = append(c.requestModifier, modifier)
}
