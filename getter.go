package librediscovery

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
	"github.com/jiuzhou-zhao/go-fundamental/loge"
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
	opts           discovery.Options
	ob             discovery.Observer

	cachedSs []*discovery.ServiceInfo
}

func NewGetter(ctx context.Context, logger interfaces.Logger, redisCli *redis.Client, key string,
	expireDuration time.Duration, checkDuration time.Duration) (discovery.Getter, error) {
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
	return &getterServerImpl{
		CycleServiceWrapper: service_wrapper.NewCycleServiceWrapper(ctx, logger),
		redisCli:            redisCli,
		key:                 discoveryKeyPre + ":" + key,
		expireDuration:      expireDuration,
		checkDuration:       checkDuration,
		cachedSs:            make([]*discovery.ServiceInfo, 0),
	}, nil
}

func (getter *getterServerImpl) unmarshalAndCheckRedisInfo(d []byte) (info *redisInfo4DiscoveryWithTouchTm, err error) {
	var i redisInfo4DiscoveryWithTouchTm
	err = json.Unmarshal(d, &i)
	if err != nil {
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
	latestSs := make([]*discovery.ServiceInfo, 0)

	eLog := loge.NewLogger(logger)

	var cursor uint64
	var keys []string
	var err error
	for {
		utils.DefRedisTimeoutOpEx(ctx, func(ctx context.Context) {
			keys, cursor, err = getter.redisCli.HScan(ctx, getter.key, cursor, getter.opts.String(), 10).Result()
		})
		if err != nil {
			eLog.Errorf(ctx, "redis failed: %v", err)
			return
		}

		for idx := 0; idx < len(keys); idx += 2 {
			info, err := getter.unmarshalAndCheckRedisInfo([]byte(keys[idx+1]))
			if err != nil {
				eLog.Errorf(ctx, "parse discovery item failed: %v, %v", err, keys[idx+1])
				utils.DefRedisTimeoutOp(func(ctx context.Context) {
					err = getter.redisCli.HDel(ctx, getter.key, keys[idx]).Err()
					if err != nil {
						eLog.Errorf(ctx, "redis failed: %v", err)
					}
				})
				continue
			}
			latestSs = append(latestSs, info.ServiceInfo)
		}

		if cursor <= 0 {
			break
		}
	}

	if reflect.DeepEqual(latestSs, getter.cachedSs) {
		loge.Debug(ctx, "same server list")
		return
	}

	getter.cachedSs = latestSs
	getter.ob(getter.cachedSs)
}

func (getter *getterServerImpl) DoJob(ctx context.Context, logger interfaces.Logger) (time.Duration, error) {
	getter.doJob(ctx, logger)
	return getter.checkDuration, nil
}

func (getter *getterServerImpl) Start(ob discovery.Observer, opt ...discovery.Option) error {
	getter.opts = defaultServerOptions
	for _, o := range opt {
		o.Apply(&getter.opts)
	}
	getter.ob = ob
	return getter.CycleServiceWrapper.Start(getter)
}

func (getter *getterServerImpl) Stop() {
	getter.CycleServiceWrapper.Stop()
}

func (getter *getterServerImpl) Wait() {
	getter.CycleServiceWrapper.Wait()
}

var defaultServerOptions = discovery.Options{
	ServiceType: "*",
	ServiceName: "*",
}
