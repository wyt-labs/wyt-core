package base

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"github.com/wyt-labs/wyt-core/internal/pkg/config"
	"github.com/wyt-labs/wyt-core/pkg/basic"
	"github.com/wyt-labs/wyt-core/pkg/cache"
)

func init() {
	basic.RegisterComponents(NewBaseComponent)
}

type cronLoggerWrapper struct {
	Logger *logrus.Logger
}

func (c *cronLoggerWrapper) Info(msg string, keysAndValues ...any) {
	c.Logger.WithField("module", "cron").Info(append([]any{msg}, keysAndValues...))
}

func (c *cronLoggerWrapper) Error(err error, msg string, keysAndValues ...any) {
	c.Logger.WithFields(logrus.Fields{"err": err, "module": "cron"}).Error(append([]any{msg}, keysAndValues...))
}

type Component struct {
	*basic.BaseComponent
	Config   *config.Config
	MemCache *cache.MemCache
	Cron     *cron.Cron
}

func NewBaseComponent(baseComponent *basic.BaseComponent, config *config.Config) (*Component, error) {
	memCache, err := cache.NewMemCache(config.Cache.ExpiredTime.ToDuration(), config.Cache.CleanupInterval.ToDuration())
	if err != nil {
		return nil, err
	}
	return &Component{
		BaseComponent: baseComponent,
		Config:        config,
		MemCache:      memCache,
		Cron:          cron.New(cron.WithLogger(&cronLoggerWrapper{Logger: baseComponent.Logger})),
	}, nil
}
