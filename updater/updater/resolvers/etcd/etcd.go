package etcd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/client"
)

const ResolverName = "etcd"

func New(cfg map[string]string) (*Resolver, error) {
	headerTimeout, err := strconv.Atoi(cfg["header_timeout"])
	if err != nil {
		return nil, fmt.Errorf("failed to transform dial_timeout as string to integer: %s", err)
	}

	clientsConfig := client.Config{
		Endpoints:               strings.Split(cfg["endpoints"], ","),
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Duration(headerTimeout) * time.Second,
	}
	etcdClient, err := client.New(clientsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client for etcd: %s", err)
	}

	return &Resolver{
		kv: client.NewKeysAPI(etcdClient),
	}, nil
}

type Resolver struct {
	kv client.KeysAPI
}

func (r *Resolver) Resolve(ctx context.Context, key string) (string, error) {
	resp, err := r.kv.Get(ctx, key, nil)
	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func (r *Resolver) Name() string {
	return ResolverName
}

func (r *Resolver) Prefix() string {
	return ResolverName
}
