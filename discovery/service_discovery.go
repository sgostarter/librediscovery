package discovery

import "strings"

const (
	TypeGRpc = "grpc"
	TypeHttp = "http"
)

func IsValidType(t string) bool {
	switch strings.ToLower(t) {
	case TypeGRpc:
	case TypeHttp:
	default:
		return false
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
