package librediscovery

import (
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
)

const (
	discoveryKeyPre = "rediscovery"
)

type redisInfo4DiscoveryWithTouchTm struct {
	*discovery.ServiceInfo
	TouchTimestamp int64
}

func redisKey4DiscoveryPool(poolKey string) string {
	if poolKey == "" {
		return discoveryKeyPre
	}
	return discoveryKeyPre + ":" + poolKey
}
