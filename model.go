package librediscovery

import (
	"time"

	"github.com/sgostarter/librediscovery/discovery"
)

const (
	discoveryKeyPre = "rediscovery"
)

type redisInfo4DiscoveryWithTouchTm struct {
	*discovery.ServiceInfo
	TouchTimestamp int64
	TouchTime      time.Time
}

func redisKey4DiscoveryPool(poolKey string) string {
	if poolKey == "" {
		return discoveryKeyPre
	}
	return discoveryKeyPre + ":" + poolKey
}
