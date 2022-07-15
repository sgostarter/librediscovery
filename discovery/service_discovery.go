package discovery

import (
	"strings"
	"sync"
	"time"
)

const (
	TypeBuildInGRPC    = "@grpc"
	TypeBuildInGRPCWeb = "@grpc_web"
	TypeBuildInHTTP    = "@http"
)

var _registeredTypes sync.Map

func RegisterType(t string) {
	_registeredTypes.Store(strings.ToLower(t), time.Now())
}

func IsValidType(t string) bool {
	t = strings.ToLower(t)

	switch t {
	case TypeBuildInGRPC:
	case TypeBuildInGRPCWeb:
	case TypeBuildInHTTP:
	default:
		if _, ok := _registeredTypes.Load(t); !ok {
			return false
		}
	}

	return true
}

const (
	MetaGRPCClass = "grpc_class"
)

type ServiceInfo struct {
	Host        string
	Port        int
	ServiceName string // type:name:index
	Meta        map[string]string
}

type Observer func(services []*ServiceInfo)

type Setter interface {
	Start([]*ServiceInfo) error
	Stop()
	Wait()
}

type Getter interface {
	Start(ob Observer, opt ...Option) error
	Stop()
	Wait()
}
