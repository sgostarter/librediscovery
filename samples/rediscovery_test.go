package samples

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jiuzhou-zhao/go-fundamental/clienttoolset"
	"github.com/jiuzhou-zhao/go-fundamental/grpce"
	"github.com/jiuzhou-zhao/go-fundamental/loge"
	"github.com/jiuzhou-zhao/go-fundamental/servicetoolset"
	"github.com/sgostarter/librediscovery"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

type TestHelloWorld struct {
	helloworld.UnimplementedGreeterServer

	id string
}

func (o *TestHelloWorld) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: fmt.Sprintf("Hi %v, I'm %v", req.Name, o.id),
	}, nil
}

func Test(t *testing.T) {
	serverName := "testsvr"

	options, err := redis.ParseURL("redis://:redis_default_pass1@dev.env:8900/2")
	assert.Nil(t, err)
	redisClient := redis.NewClient(options)
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	logger := &loge.ConsoleLogger{}

	getter, err := librediscovery.NewGetter(ctx, logger, redisClient, "", 10*time.Second, time.Second)
	assert.Nil(t, err)

	go func() {
		setter, err := librediscovery.NewSetter(ctx, logger, redisClient, "", time.Second)
		assert.Nil(t, err)

		serviceToolset := servicetoolset.NewServerToolset(ctx, logger)
		err = serviceToolset.CreateGRpcServer(&servicetoolset.GRpcServerConfig{
			Name:          serverName + ":1",
			Address:       ":9001",
			DisableLog:    false,
			MetaTransKeys: nil,
			DiscoveryExConfig: servicetoolset.DiscoveryExConfig{
				Setter:          setter,
				ExternalAddress: "127.0.0.1",
			},
		}, nil, func(server *grpc.Server) {
			helloworld.RegisterGreeterServer(server, &TestHelloWorld{
				id: "node1",
			})
		})

		assert.Nil(t, err)
		err = serviceToolset.Start()
		assert.Nil(t, err)
		serviceToolset.Wait()
	}()

	go func() {
		setter, err := librediscovery.NewSetter(ctx, logger, redisClient, "", time.Second)
		assert.Nil(t, err)

		serviceToolset := servicetoolset.NewServerToolset(ctx, logger)
		err = serviceToolset.CreateGRpcServer(&servicetoolset.GRpcServerConfig{
			Name:          serverName + ":2",
			Address:       ":9002",
			DisableLog:    false,
			MetaTransKeys: nil,
			DiscoveryExConfig: servicetoolset.DiscoveryExConfig{
				Setter:          setter,
				ExternalAddress: "127.0.0.1",
			},
		}, nil, func(server *grpc.Server) {
			helloworld.RegisterGreeterServer(server, &TestHelloWorld{
				id: "node2",
			})
		})

		assert.Nil(t, err)
		err = serviceToolset.Start()
		assert.Nil(t, err)
		serviceToolset.Wait()
	}()

	schema := "rediscoverytest"

	err = grpce.RegisterResolver(getter, logger, schema)
	assert.Nil(t, err)

	conn, err := clienttoolset.DialGRpcServer(&clienttoolset.GRpcClientConfig{
		Address: fmt.Sprintf("%s:///%s", schema, serverName),
	}, []grpc.DialOption{grpc.WithDefaultServiceConfig(`
{
	"loadBalancingConfig": [ { "round_robin": {} } ]
}
`)})
	assert.Nil(t, err)

	cli := helloworld.NewGreeterClient(conn)
	for idx := 0; idx < 10; idx++ {
		time.Sleep(time.Second)
		resp, err := cli.SayHello(context.Background(), &helloworld.HelloRequest{
			Name: "tester1",
		})
		if err != nil {
			t.Logf("error: %v", err)
			continue
		}
		t.Log(resp.Message)
	}
}
