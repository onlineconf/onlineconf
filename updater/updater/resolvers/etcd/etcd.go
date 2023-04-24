package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/rs/zerolog/log"
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
		log.Error().Msgf("etcd get error: %s", err)
		return "", err
	}

	if len(resp.Kvs) == 0 {
		log.Error().Msgf("etcd empty kvs")
		return "", nil
	}

	return string(resp.Kvs[0].Value), nil
}

func (r *Resolver) Name() string {
	return ResolverName
}

func (r *Resolver) Prefix() string {
	return ResolverName
}
