package librediscovery

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
	"github.com/stretchr/testify/assert"
)

func TestGetterSetter(t *testing.T) {
	options, err := redis.ParseURL("redis://:redis_default_pass1@dev.env:8900/2")
	assert.Nil(t, err)
	redisClient := redis.NewClient(options)
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	setter, err := NewSetter(ctx, nil, redisClient, "", time.Second)
	assert.Nil(t, err)
	err = setter.Start([]*discovery.ServiceInfo{
		{
			Host:        "127.0.0.1",
			Port:        1000,
			ServiceName: discovery.BuildDiscoveryServerName("grpc", "server1", "1"),
			Meta: map[string]string{
				"a": "b",
			},
		},
		{
			Host:        "127.0.0.1",
			Port:        1001,
			ServiceName: discovery.BuildDiscoveryServerName("grpc", "server1", "2"),
			Meta: map[string]string{
				"d": "c",
			},
		},
		{
			Host:        "127.0.0.1",
			Port:        1002,
			ServiceName: discovery.BuildDiscoveryServerName("http", "server1", "2"),
			Meta: map[string]string{
				"d": "c",
			},
		},
		{
			Host:        "127.0.0.1",
			Port:        1002,
			ServiceName: discovery.BuildDiscoveryServerName("http", "server2", "2"),
			Meta: map[string]string{
				"d": "c",
			},
		},
		{
			Host:        "127.0.0.1",
			Port:        10032,
			ServiceName: "原始hahahoho",
			Meta: map[string]string{
				"d": "c",
			},
		},
	})
	assert.Nil(t, err)

	fnSerivicesDump := func(id string, services []*discovery.ServiceInfo) {
		ss := strings.Builder{}
		ss.WriteString(fmt.Sprintf("<<<<<<< %v =====\n", id))
		for _, service := range services {
			ss.WriteString(fmt.Sprintf("%v %v %v %v\n", service.ServiceName, service.Host, service.Port, service.Meta))
		}
		ss.WriteString(fmt.Sprintf(">>>>>>> %v =====\n", id))
		t.Log(ss.String())
	}

	getter, err := NewGetter(ctx, nil, redisClient, "", 5*time.Second, 2*time.Second)
	assert.Nil(t, err)
	err = getter.Start(func(services []*discovery.ServiceInfo) {
		fnSerivicesDump("all", services)
	})
	assert.Nil(t, err)

	getter, err = NewGetter(ctx, nil, redisClient, "", 5*time.Second, 2*time.Second)
	assert.Nil(t, err)
	err = getter.Start(func(services []*discovery.ServiceInfo) {
		fnSerivicesDump("grpc", services)
	}, discovery.TypeOption("grpc"))
	assert.Nil(t, err)

	getter, err = NewGetter(ctx, nil, redisClient, "", 5*time.Second, 2*time.Second)
	assert.Nil(t, err)
	err = getter.Start(func(services []*discovery.ServiceInfo) {
		fnSerivicesDump("server1", services)
	}, discovery.NameOption("server1"))
	assert.Nil(t, err)

	getter, err = NewGetter(ctx, nil, redisClient, "", 5*time.Second, 2*time.Second)
	assert.Nil(t, err)
	err = getter.Start(func(services []*discovery.ServiceInfo) {
		fnSerivicesDump("原始hahahoho", services)
	}, discovery.RawKeyOption("原始hahahoho"))
	assert.Nil(t, err)

	setter.Wait()
	getter.Wait()
}
