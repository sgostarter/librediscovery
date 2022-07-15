package discovery

import (
	"fmt"
	"strings"
)

func BuildDiscoveryServerName(t string, name string, index string) string {
	t = strings.ToLower(t)

	serverName := fmt.Sprintf("%s:%s", t, name)
	if index != "" {
		serverName = serverName + ":" + index
	}

	return serverName
}

func ParseDiscoveryServerName(n string) (t string, name string, index string, err error) {
	vs := strings.Split(n, ":")
	if len(vs) < 2 {
		err = fmt.Errorf("invalid server name: %v", n)

		return
	}

	t = vs[0]
	t = strings.ToLower(t)

	if !IsValidType(t) {
		err = fmt.Errorf("unknown type: %v", t)

		return
	}

	switch len(vs) {
	case 2:
		name = vs[1]
	case 3:
		name = vs[1]
		index = vs[2]
	default:
		name = strings.Join(vs[1:len(vs)-1], ":")
		index = vs[len(vs)-1]
	}

	return
}
