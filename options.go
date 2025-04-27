package playwrightcigo

import (
	"context"
	"time"
)

type Option interface {
	apply(*config)
}

var _ Option = (*optionFunc)(nil)

type optionFunc func(*config)

func (f optionFunc) apply(c *config) {
	f(c)
}

func WithContext(ctx context.Context) Option {
	return optionFunc(func(c *config) {
		c.ctx = ctx
	})
}

func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(c *config) {
		c.timeout = timeout
	})
}

func WithRepository(repository, tag string) Option {
	return optionFunc(func(c *config) {
		if repository != "" {
			c.repository = repository
		}
		if tag != "" {
			c.tag = tag
		}
	})
}
