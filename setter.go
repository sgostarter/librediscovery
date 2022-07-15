package librediscovery

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libeasygo/helper"
	"github.com/sgostarter/libeasygo/servicewrapper"
	"github.com/sgostarter/librediscovery/discovery"
)

type setterServerImpl struct {
	*servicewrapper.CycleServiceWrapper

	ctx            context.Context
	logger         l.Wrapper
	redisCli       *redis.Client
	poolKey        string
	updateDuration time.Duration

	services []*redisInfo4DiscoveryWithTouchTm
}

func NewDefaultSetter(ctx context.Context, redisCli *redis.Client) (discovery.Setter, error) {
	return NewSetter(ctx, l.NewNopLoggerWrapper(), redisCli, "", time.Minute)
}

func NewSetter(ctx context.Context, logger l.Wrapper, redisCli *redis.Client, poolKey string,
	updateDuration time.Duration) (discovery.Setter, error) {
	if logger == nil {
		logger = l.NewNopLoggerWrapper()
	}

	if redisCli == nil {
		return nil, errors.New("no redis")
	}

	if updateDuration <= 0 {
		return nil, errors.New("no update duration")
	}

	return &setterServerImpl{
		CycleServiceWrapper: servicewrapper.NewCycleServiceWrapper(ctx, logger),
		ctx:                 ctx,
		logger:              logger.WithFields(l.StringField(l.ClsKey, "setter")),
		redisCli:            redisCli,
		poolKey:             redisKey4DiscoveryPool(poolKey),
		updateDuration:      updateDuration,
		services:            nil,
	}, nil
}

func (setter *setterServerImpl) DoJob(ctx context.Context, logger l.Wrapper) (time.Duration, error) {
	for idx := range setter.services {
		setter.services[idx].TouchTime = time.Now()
		setter.services[idx].TouchTimestamp = setter.services[idx].TouchTime.Unix()
		bs, _ := json.Marshal(setter.services[idx])
		helper.DoWithTimeout(ctx, time.Second, func(ctx context.Context) {
			err := setter.redisCli.HSet(ctx, setter.poolKey, setter.services[idx].ServiceName, bs).Err()
			if err != nil {
				logger.Errorf("publish service %v failed: %v",
					setter.services[idx].ServiceName, err)
			}
		})
	}
	return setter.updateDuration, nil
}

func (setter *setterServerImpl) Start(services []*discovery.ServiceInfo) error {
	if len(setter.services) > 0 {
		setter.logger.Fatal(setter.ctx, "try overflow the services settings")
		return nil
	}
	serverInfos := make([]*redisInfo4DiscoveryWithTouchTm, 0, len(services))
	for _, service := range services {
		serverInfos = append(serverInfos, &redisInfo4DiscoveryWithTouchTm{
			ServiceInfo:    service,
			TouchTimestamp: 0,
		})
	}
	err := setter.CycleServiceWrapper.Start(setter)
	if err != nil {
		return err
	}
	setter.services = serverInfos
	return nil
}

func (setter *setterServerImpl) Stop() {
	setter.CycleServiceWrapper.Stop()
}

func (setter *setterServerImpl) Wait() {
	setter.CycleServiceWrapper.Wait()
}
