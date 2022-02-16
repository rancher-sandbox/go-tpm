package tpm

import "github.com/google/go-attestation/attest"

type Config struct {
	emulated       bool
	commandChannel attest.CommandChannelTPM20
	seed           int64
}

type Option func(c *Config) error

var Emulated Option = func(c *Config) error {
	c.emulated = true
	return nil
}

func WithSeed(s int64) Option {
	return func(c *Config) error {
		c.seed = s
		return nil
	}
}

func WithCommandChannel(cc attest.CommandChannelTPM20) Option {
	return func(c *Config) error {
		c.commandChannel = cc
		return nil
	}
}

func (c *Config) Apply(opts ...Option) error {
	for _, o := range opts {
		if err := o(c); err != nil {
			return err
		}
	}

	return nil
}
