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
	TouchTimestamp int64     `json:"touch_timestamp"`
	TouchTime      time.Time `json:"touch_time"`
}

func redisKey4DiscoveryPool(poolKey string) string {
	if poolKey == "" {
		return discoveryKeyPre
	}

	return discoveryKeyPre + ":" + poolKey
}
