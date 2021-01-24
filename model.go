package librediscovery

import (
	"github.com/jiuzhou-zhao/go-fundamental/clienttoolset"
)

type ServiceInfo4Discovery struct {
	Name             string
	Path             string
	GRpcClientConfig *clienttoolset.GRpcClientConfig
}

type redisInfo4DiscoveryWithTouchTm struct {
	*ServiceInfo4Discovery
	TouchTimestamp int64
}
