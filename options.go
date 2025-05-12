package playwrightcigo

import (
	"context"
	"time"
)

// Option is an interface for configuring the behavior of playwright-ci-go functions.
type Option interface {
	apply(*config)
}

var _ Option = (*optionFunc)(nil)

type optionFunc func(*config)

func (f optionFunc) apply(c *config) {
	f(c)
}

// WithContext provides a context for cancellation and timeout control.
// This option allows you to control when operations should be canceled.
func WithContext(ctx context.Context) Option {
	return optionFunc(func(c *config) {
		c.ctx = ctx
	})
}

// WithTimeout sets the timeout of the container.
// The default timeout is 5 minutes.
func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(c *config) {
		c.timeout = timeout
	})
}

// WithRetry configures the number of retry attempts when setting up the proxy.
// The default is 15 retries.
func WithRetry(count int) Option {
	return optionFunc(func(c *config) {
		if count > 0 {
			c.retry = count
		}
	})
}

// WithSleeping sets the sleep duration between retry attempts when setting up the proxy.
// The default is 200 milliseconds.
func WithSleeping(sleeping time.Duration) Option {
	return optionFunc(func(c *config) {
		if sleeping > 0 {
			c.sleeping = sleeping
		}
	})
}

// WithRepository sets a custom container repository and tag.
// This is useful for using custom container images or specific versions.
// The default repository is "ghcr.io/mountain-reverie/playwright-ci-go".
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
