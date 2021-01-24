package rediscovery

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/clienttoolset"
	"github.com/jiuzhou-zhao/go-fundamental/discovery"
	"github.com/stretchr/testify/assert"
)

func TestGetterSetter(t *testing.T) {
	options, err := redis.ParseURL("redis://:redis_default_pass1@dev.env:8900/3")
	assert.Nil(t, err)
	redisClient := redis.NewClient(options)
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	setter, err := NewSetter(ctx, nil, redisClient, "", time.Second, []*ServiceInfo4Discovery{
		{
			Name: "name1",
			Path: "/aa/bb",
			GRpcClientConfig: &clienttoolset.GRpcClientConfig{
				Address: "127.0.0.1:1000",
			},
		},
		{
			Name: "name2",
			Path: "/aa/bb2",
			GRpcClientConfig: &clienttoolset.GRpcClientConfig{
				Address: "127.0.0.2:1000",
			},
		},
	})
	assert.Nil(t, err)
	err = setter.Start()
	assert.Nil(t, err)

	getter, err := NewGetter(ctx, nil, redisClient, "", 5*time.Second, 2*time.Second, func(services map[string][]*discovery.ServiceInfo) {
		t.Logf("begin ...\n")
		for k, infos := range services {
			t.Logf("key=%v\n", k)
			for _, info := range infos {
				t.Logf("  %v, %v\n", info.Name, info.GRpcAddresses.Address)
			}
		}
		t.Logf("finish ...\n")
	})
	assert.Nil(t, err)
	err = getter.Start()
	assert.Nil(t, err)

	setter.Wait()
	getter.Wait()
}
