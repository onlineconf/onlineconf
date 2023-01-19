package updater

import (
	"context"
	"github.com/onlineconf/onlineconf/updater/v3/updater/resolvers/etcd"
	"github.com/rs/zerolog/log"
	"strings"
)

type IResolverPlugin interface {
	Resolve(context.Context, string) (string, error)
	Info() string
	Prefix() string
	Name() string
}

func InitPluginsUsage(pluginsCfgs map[string]map[string]string) {
	etcd, err := etcd.New(pluginsCfgs[etcd.ResolverName])
	if err == nil {
		IncludedPlugins = append(IncludedPlugins, etcd)
	} else {
		log.Error().Err(err).Msg("cant init " + etcd.Name() + "plugin")
	}
}

var IncludedPlugins []IResolverPlugin

func TryResolveByResolverPlugins(ctx context.Context, key string) (resolved string, ok bool) {
	for _, p := range IncludedPlugins {
		if strings.HasPrefix(key, p.Prefix()) {
			val, err := p.Resolve(ctx, strings.TrimPrefix(key, p.Prefix()))
			if err != nil {
				return "", false
			}

			return val, true
		}
	}

	return "", false
}
