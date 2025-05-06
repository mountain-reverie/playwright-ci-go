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

func WithRetry(count int) Option {
	return optionFunc(func(c *config) {
		if count > 0 {
			c.retry = count
		}
	})
}

func WithSleeping(sleeping time.Duration) Option {
	return optionFunc(func(c *config) {
		if sleeping > 0 {
			c.sleeping = sleeping
		}
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
