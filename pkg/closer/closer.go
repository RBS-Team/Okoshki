package closer

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/RBS-Team/Okoshki/pkg/logger"
)

type closeFn struct {
	name string
	fn   func(context.Context) error
}

type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	funcs  []closeFn
	logger logger.Logger
}

func New(log logger.Logger) *Closer {
	return &Closer{logger: log}
}

func (c *Closer) Add(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, closeFn{name: name, fn: fn})
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			return
		}

		var errs []error

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			start := time.Now()
			c.logger.Infof("closing %s...", f.name)

			if err := f.fn(ctx); err != nil {
				c.logger.Errorf("failed to close %s: %v (elapsed: %s)", f.name, err, time.Since(start))
				errs = append(errs, err)
			} else {
				c.logger.Infof("%s closed (elapsed: %s)", f.name, time.Since(start))
			}
		}

		result = errors.Join(errs...)
	})

	return result
}
