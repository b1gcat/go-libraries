package iot

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

type Iot interface {
	CreateThing(map[string]interface{}) (map[string]interface{}, error)
	Subscribe(sub map[string]interface{}) error
	Publish(pub map[string]interface{}) error
	Config() map[string]interface{}
	Close()
}

type Options struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Logger
	once   sync.Once
}

type OptionsFunc func(*Options) error

func defaultOption() Options {
	opt := Options{
		logger: logrus.New(),
	}
	opt.ctx, opt.cancel = context.WithCancel(context.Background())
	return opt
}

func WithContext(ctx context.Context) OptionsFunc {
	return func(o *Options) error {
		o.ctx, o.cancel = context.WithCancel(ctx)
		return nil
	}
}

func WithLogger(l *logrus.Logger) OptionsFunc {
	return func(o *Options) error {
		o.logger = l
		return nil
	}
}
