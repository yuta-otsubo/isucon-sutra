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
	agent *agent.agent

	contestantLogger *zap.Logger
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
		agent.WithCache(),
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
