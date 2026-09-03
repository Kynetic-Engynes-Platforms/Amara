package connection

import (
	"net"
	"net/http"
	"time"

	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/augment"
	"github.com/Kynetic-Engynes-Platforms/typesense-go/pkg/impls/types"
)

func NewClient(cfg types.Config) (*types.Client, error) {
	if len(cfg.Nodes) == 0 {
		cfg.Nodes = []string{"http://localhost:8108"}
	}
	if cfg.NumRetries == 0 {
		cfg.NumRetries = 3
	}
	if cfg.HealthWaitTime == 0 {
		cfg.HealthWaitTime = 60 * time.Second
	}
	if cfg.RetryInterval == 0 {
		cfg.RetryInterval = 100 * time.Millisecond
	}

	return augment.AugmentedClient(&types.Client{
		Config: cfg,
		Http: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 5 * time.Second,
			},
			Timeout: 15 * time.Second,
		},
	}, cfg)

}
