package updater

import (
	"github.com/onlineconf/onlineconf/updater/v3/updater/resolver-plugins/etcd"
	"github.com/rs/zerolog/log"
	"strings"
)

type IResolverPlugin interface {
	Resolve(string) (string, error)
	Info() string
	Prefix() string
	Name() string
}

func InitPluginsUsage() {
	etcd, err := etcd.New()
	if err == nil {
		IncludedPlugins = append(IncludedPlugins, etcd)
	} else {
		log.Error().Err(err).Msg("cant init " + etcd.Name() + "plugin")
	}
}

var IncludedPlugins []IResolverPlugin

func TryResolveByResolverPlugins(key string) (resolved string, ok bool) {
	for _, p := range IncludedPlugins {
		if strings.HasPrefix(key, p.Prefix()) {
			val, err := p.Resolve(strings.TrimPrefix(key, p.Prefix()))
			if err != nil {
				return "", false
			}

			return val, true
		}
	}

	return "", false
}
