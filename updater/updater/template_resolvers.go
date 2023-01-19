package updater

import (
	"context"
	"github.com/onlineconf/onlineconf/updater/v3/updater/resolvers/etcd"
	"github.com/rs/zerolog/log"
	"strings"
)

type ResolverModule interface {
	Resolve(context.Context, string) (string, error)
	Prefix() string
	Name() string
}

func InitResolveModulesUsage(resolversCfgs map[string]map[string]string) {
	etcd, err := etcd.New(resolversCfgs[etcd.ResolverName])
	if err != nil {
		log.Error().Err(err).Msg("cant init " + etcd.Name() + "resolver module")
	} else {
		IncludedResolveModules = append(IncludedResolveModules, etcd)
	}
}

var IncludedResolveModules []ResolverModule

func TryResolveByResolverModule(ctx context.Context, key string) (resolved string, ok bool) {
	for _, m := range IncludedResolveModules {
		if strings.HasPrefix(key, m.Prefix()) {
			val, err := m.Resolve(ctx, strings.TrimPrefix(key, m.Prefix()))
			if err != nil {
				return "", false
			}

			return val, true
		}
	}

	return "", false
}
