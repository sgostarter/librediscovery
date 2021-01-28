package librediscovery

import (
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
)

type redisInfo4DiscoveryWithTouchTm struct {
	*discovery.ServiceInfo
	TouchTimestamp int64
}
