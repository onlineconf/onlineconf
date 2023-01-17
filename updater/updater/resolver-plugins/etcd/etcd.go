package etcd

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"strings"
	"time"
)

const ResolverName = "etcd"

func New(cfg map[string]string) (*Resolver, error) {
	dialTimeout, err := time.ParseDuration(cfg["dial_timeout"])
	if err != nil {
		return nil, err
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(cfg["endpoints"], ","),
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}

	kv := clientv3.NewKV(cli)

	return &Resolver{
		kv: kv,
	}, nil
}

type Resolver struct {
	kv clientv3.KV
}

func (r *Resolver) Resolve(ctx context.Context, key string) (string, error) {
	return "", nil
	resp, err := r.kv.Get(ctx, key)
	if err != nil {
		return "", err
	}

	fmt.Println(resp)
	return "etcd", nil
}

func (r *Resolver) Info() string {
	return ""
}

func (r *Resolver) Name() string {
	return ResolverName
}

func (r *Resolver) Prefix() string {
	return ResolverName + ":"
}
