package updater

import (
	"context"
	"fmt"
	"strings"

	"github.com/onlineconf/onlineconf/updater/v3/updater/resolvers/etcd"
)

type TemplateVariableResolver interface {
	Resolve(context.Context, string) (string, error)
	Prefix() string
	Name() string
}

func InitResolveModulesUsage(resolversCfgs map[string]map[string]string) error {
	etcd, err := etcd.New(resolversCfgs[etcd.ResolverName])
	if err != nil {
		return fmt.Errorf("cant init etcd resolve module: %w", err)
	}
	IncludedResolveModules = append(IncludedResolveModules, etcd)

	return nil
}

var IncludedResolveModules []TemplateVariableResolver

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
