package kwok

import (
	"context"
	"log/slog"

	"github.com/dejanzele/batch-simulator/internal/cmd"
)

const (
	// LabelSelector which identifies resources managed by kwok.
	LabelSelector = "type=kwok"
)

type Client struct {
	logger *slog.Logger
}

type Option func(*Client)

func WithLogger(logger *slog.Logger) Option {
	return func(k *Client) {
		k.logger = logger
	}
}

func New(opts ...Option) *Client {
	k := &Client{}
	for _, opt := range opts {
		opt(k)
	}
	if k.logger == nil {
		k.logger = &slog.Logger{}
	}
	k.logger = k.logger.With("process", "kwok")
	return k
}

func (c *Client) Run(ctx context.Context) error {
	runner := cmd.New(c.logger)
	_, err := runner.Run(ctx, "kwok", DefaultArgs()...)
	return err
}

func DefaultArgs() []string {
	return []string{
		"--kubeconfig=~/.kube/config",
		"--manage-all-nodes=false",
		"--manage-nodes-with-annotation-selector=kwok.x-k8s.io/node=fake",
		"--manage-nodes-with-label-selector=",
		"--manage-single-node=",
		"--disregard-status-with-annotation-selector=kwok.x-k8s.io/status=custom",
		"--disregard-status-with-label-selector=",
		"--cidr=10.0.0.1/24",
		"--node-ip=10.0.0.1",
		"--node-lease-duration-seconds=40",
		"--disable-qps-limit",
		"--enable-crds=Stage",
	}
}
