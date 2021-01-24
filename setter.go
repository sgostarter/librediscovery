package librediscovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/interfaces"
	"github.com/jiuzhou-zhao/go-fundamental/serviceutils/service_wrapper"
	"github.com/jiuzhou-zhao/go-fundamental/utils"
)

type setterServerImpl struct {
	*service_wrapper.CycleServiceWrapper

	redisCli       *redis.Client
	poolKey        string
	updateDuration time.Duration

	services []*redisInfo4DiscoveryWithTouchTm
}

func NewSetter(ctx context.Context, logger interfaces.Logger, redisCli *redis.Client, poolKey string,
	updateDuration time.Duration, services []*ServiceInfo4Discovery) (interfaces.Server, error) {
	if redisCli == nil {
		return nil, errors.New("no redis")
	}

	if updateDuration <= 0 {
		return nil, errors.New("no update duration")
	}

	if len(services) <= 0 {
		return nil, errors.New("no service")
	}

	for _, service := range services {
		if service == nil || service.Name == "" || !strings.HasPrefix(service.Path, "/") ||
			service.GRpcClientConfig == nil || service.GRpcClientConfig.Address == "" {
			return nil, fmt.Errorf("invalid service: %v", service)
		}
	}

	swtm := make([]*redisInfo4DiscoveryWithTouchTm, 0, len(services))
	for idx := range services {
		swtm = append(swtm, &redisInfo4DiscoveryWithTouchTm{
			ServiceInfo4Discovery: services[idx],
			TouchTimestamp:        0,
		})
	}
	return &setterServerImpl{
		CycleServiceWrapper: service_wrapper.NewCycleServiceWrapper(ctx, logger),
		redisCli:            redisCli,
		poolKey:             discoveryKeyPre + ":" + poolKey,
		updateDuration:      updateDuration,
		services:            swtm,
	}, nil
}

func (setter *setterServerImpl) DoJob(ctx context.Context, logger interfaces.Logger) (time.Duration, error) {
	for idx := range setter.services {
		setter.services[idx].TouchTimestamp = time.Now().Unix()
		bs, _ := json.Marshal(setter.services[idx])
		utils.DefRedisTimeoutOpEx(ctx, func(ctx context.Context) {
			err := setter.redisCli.HSet(ctx, setter.poolKey, setter.services[idx].Name, bs).Err()
			if err != nil {
				logger.Recordf(ctx, interfaces.LogLevelError, "publish service %v failed: %v",
					setter.services[idx].Name, err)
			}
		})
	}
	return setter.updateDuration, nil
}

func (setter *setterServerImpl) Start() error {
	return setter.CycleServiceWrapper.Start(setter)
}

func (setter *setterServerImpl) Stop() {
	setter.CycleServiceWrapper.Stop()
}

func (setter *setterServerImpl) Wait() {
	setter.CycleServiceWrapper.Wait()
}
