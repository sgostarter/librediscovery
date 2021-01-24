package rediscovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
	"github.com/jiuzhou-zhao/go-fundamental/interfaces"
	"github.com/jiuzhou-zhao/go-fundamental/serviceutils/service_wrapper"
	"github.com/jiuzhou-zhao/go-fundamental/utils"
)

const (
	discoveryKeyPre = "rediscovery"
)

type getterServerImpl struct {
	*service_wrapper.CycleServiceWrapper
	redisCli       *redis.Client
	key            string
	expireDuration time.Duration
	checkDuration  time.Duration
	ob             discovery.Observer

	cachedSs map[string][]*discovery.ServiceInfo
}

func NewGetter(ctx context.Context, logger interfaces.Logger, redisCli *redis.Client, key string,
	expireDuration time.Duration, checkDuration time.Duration, ob discovery.Observer) (interfaces.Server, error) {
	if redisCli == nil {
		return nil, errors.New("no redis")
	}
	if expireDuration <= 0 {
		return nil, errors.New("no expire duration")
	}
	if checkDuration <= 0 {
		return nil, errors.New("no check duration")
	}
	if checkDuration >= expireDuration {
		return nil, errors.New("invalid check/expire duration")
	}
	if ob == nil {
		return nil, errors.New("no ob")
	}
	return &getterServerImpl{
		CycleServiceWrapper: service_wrapper.NewCycleServiceWrapper(ctx, logger),
		redisCli:            redisCli,
		key:                 discoveryKeyPre + ":" + key,
		expireDuration:      expireDuration,
		checkDuration:       checkDuration,
		cachedSs:            make(map[string][]*discovery.ServiceInfo),
		ob:                  ob,
	}, nil
}

func (getter *getterServerImpl) unmarshalAndCheckRedisInfo(d []byte) (info *redisInfo4DiscoveryWithTouchTm, err error) {
	var i redisInfo4DiscoveryWithTouchTm
	err = json.Unmarshal(d, &i)
	if err != nil {
		return
	}
	if i.Name == "" || i.Path == "" || i.GRpcClientConfig == nil {
		err = errors.New("invalid record")
		return
	}
	if i.TouchTimestamp > 0 {
		elapsedDuration := time.Since(time.Unix(i.TouchTimestamp, 0))
		if elapsedDuration >= getter.expireDuration {
			err = fmt.Errorf("exipre: %v", elapsedDuration-getter.expireDuration)
			return
		}
	}
	info = &i
	return
}

func (getter *getterServerImpl) doJob(ctx context.Context, logger interfaces.Logger) {
	latestSs := make(map[string][]*discovery.ServiceInfo)

	var cursor uint64
	var keys []string
	var err error
	for {
		utils.DefRedisTimeoutOpEx(ctx, func(ctx context.Context) {
			keys, cursor, err = getter.redisCli.HScan(ctx, getter.key, cursor, "*", 10).Result()
		})
		if err != nil {
			logger.Recordf(ctx, interfaces.LogLevelError, "redis failed: %v", err)
			return
		}

		for idx := 0; idx < len(keys); idx += 2 {
			info, err := getter.unmarshalAndCheckRedisInfo([]byte(keys[idx+1]))
			if err != nil {
				logger.Recordf(ctx, interfaces.LogLevelError, "parse discovery item failed: %v, %v", err, keys[idx+1])
				utils.DefRedisTimeoutOp(func(ctx context.Context) {
					err = getter.redisCli.HDel(ctx, getter.key, keys[idx]).Err()
					if err != nil {
						logger.Recordf(ctx, interfaces.LogLevelError, "redis failed: %v", err)
					}
				})
				continue
			}
			if _, ok := latestSs[info.Path]; !ok {
				latestSs[info.Path] = make([]*discovery.ServiceInfo, 0, 1)
			}
			latestSs[info.Path] = append(latestSs[info.Path], &discovery.ServiceInfo{
				Name:          info.Name,
				GRpcAddresses: info.GRpcClientConfig,
			})
		}

		if cursor <= 0 {
			break
		}
	}

	if reflect.DeepEqual(latestSs, getter.cachedSs) {
		logger.Recordf(ctx, interfaces.LogLevelDebug, "same server list")
		return
	}

	getter.cachedSs = latestSs
	getter.ob(getter.cachedSs)
}

func (getter *getterServerImpl) DoJob(ctx context.Context, logger interfaces.Logger) (time.Duration, error) {
	getter.doJob(ctx, logger)
	return getter.checkDuration, nil
}

func (getter *getterServerImpl) Start() error {
	return getter.CycleServiceWrapper.Start(getter)
}

func (getter *getterServerImpl) Stop() {
	getter.CycleServiceWrapper.Stop()
}

func (getter *getterServerImpl) Wait() {
	getter.CycleServiceWrapper.Wait()
}
