package librediscovery

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
	"github.com/jiuzhou-zhao/go-fundamental/interfaces"
	"github.com/jiuzhou-zhao/go-fundamental/loge"
	"github.com/jiuzhou-zhao/go-fundamental/serviceutils/service_wrapper"
	"github.com/jiuzhou-zhao/go-fundamental/utils"
)

type setterServerImpl struct {
	*service_wrapper.CycleServiceWrapper

	ctx            context.Context
	logger         *loge.Logger
	redisCli       *redis.Client
	poolKey        string
	updateDuration time.Duration

	services []*redisInfo4DiscoveryWithTouchTm
}

func NewSetter(ctx context.Context, logger interfaces.Logger, redisCli *redis.Client, poolKey string,
	updateDuration time.Duration) (discovery.Setter, error) {
	if redisCli == nil {
		return nil, errors.New("no redis")
	}

	if updateDuration <= 0 {
		return nil, errors.New("no update duration")
	}

	return &setterServerImpl{
		CycleServiceWrapper: service_wrapper.NewCycleServiceWrapper(ctx, logger),
		ctx:                 ctx,
		logger:              loge.NewLogger(logger),
		redisCli:            redisCli,
		poolKey:             discoveryKeyPre + ":" + poolKey,
		updateDuration:      updateDuration,
		services:            nil,
	}, nil
}

func (setter *setterServerImpl) DoJob(ctx context.Context, logger interfaces.Logger) (time.Duration, error) {
	for idx := range setter.services {
		setter.services[idx].TouchTimestamp = time.Now().Unix()
		bs, _ := json.Marshal(setter.services[idx])
		utils.DefRedisTimeoutOpEx(ctx, func(ctx context.Context) {
			err := setter.redisCli.HSet(ctx, setter.poolKey, setter.services[idx].ServiceName, bs).Err()
			if err != nil {
				logger.Recordf(ctx, interfaces.LogLevelError, "publish service %v failed: %v",
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
