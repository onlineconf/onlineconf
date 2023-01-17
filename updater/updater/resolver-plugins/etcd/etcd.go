package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
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
	resp, err := r.kv.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

func (r *Resolver) Info() string {
	return "etcd resolver plugin can resolve your template values by etcd: prefix"
}

func (r *Resolver) Name() string {
	return ResolverName
}

func (r *Resolver) Prefix() string {
	return ResolverName + ":"
}
