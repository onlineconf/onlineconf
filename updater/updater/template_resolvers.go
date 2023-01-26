package updater

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/onlineconf/onlineconf/updater/v3/updater/resolvers/etcd"
	"github.com/rs/zerolog/log"
)

const separator = ":"

type TemplateVariableResolver interface {
	Resolve(context.Context, string) (string, error)
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
var ResolveChecker *ResolveModulesValChecker

func (c *ResolveModulesValChecker) CleanStorage() {
	c.storage.clean()
}

func TryResolveByResolverModule(ctx context.Context, key string) (resolved string, ok bool) {
	for _, m := range IncludedResolveModules {
		if strings.HasPrefix(key, m.Name()+separator) { // that means that val created for resolver usage
			resolverKey := strings.TrimPrefix(key, m.Name()+separator)
			val, err := m.Resolve(ctx, resolverKey)
			ResolveChecker.Put(m, resolverKey, val)

			if err != nil {
				return "", false
			}

			return val, true
		}
	}

	return "", false
}

type kv struct {
	key   string
	value string
}

type ResolveModulesValChecker struct {
	ctx             context.Context
	rebuildTreeFunc func() error
	cancel          context.CancelFunc
	tickReady       atomic.Bool
	ticker          *time.Ticker
	storage         moduleCheckVarsStorage
}

type moduleCheckVarsStorage struct {
	pluginCheckVars map[TemplateVariableResolver][]kv // plugin - (key : current value)
	mu              sync.Mutex
}

func (s *moduleCheckVarsStorage) clean() {
	s.mu.Lock()
	s.pluginCheckVars = make(map[TemplateVariableResolver][]kv)
	s.mu.Unlock()
}

func (c *ResolveModulesValChecker) Put(p TemplateVariableResolver, key string, value string) {
	c.storage.mu.Lock()
	c.storage.pluginCheckVars[p] = append((c.storage.pluginCheckVars)[p], kv{key: key, value: value})
	c.storage.mu.Unlock()
}

func NewResolveModulesValChecker(ctx context.Context, d time.Duration, rebuildFunc func() error) *ResolveModulesValChecker {
	ctx, cancel := context.WithCancel(ctx)

	checker := &ResolveModulesValChecker{
		ctx:             ctx,
		cancel:          cancel,
		ticker:          time.NewTicker(d),
		rebuildTreeFunc: rebuildFunc,
		tickReady:       atomic.Bool{},
		storage: moduleCheckVarsStorage{
			pluginCheckVars: make(map[TemplateVariableResolver][]kv),
			mu:              sync.Mutex{},
		},
	}

	return checker
}

func (c *ResolveModulesValChecker) StartCronCheck() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.ticker.C:
			if !c.treeConsistent() {
				c.rebuildTreeFunc()
			}
		}
	}
}

func (c *ResolveModulesValChecker) treeConsistent() bool {
	c.storage.mu.Lock()
	defer c.storage.mu.Unlock()

	for plugin, kvs := range c.storage.pluginCheckVars {
		for _, kv := range kvs {
			currentVal, err := plugin.Resolve(c.ctx, kv.key)
			if err != nil {
				log.Error().Err(err).
					Str("plugin", plugin.Name()).
					Msg("cant check consistency of value by resolve module")
			}

			previousValue := kv.value

			if previousValue != currentVal {
				log.Info().Msg("consistency of value by resolve module; go to rebuild tree")
				return false
			}
		}
	}

	return true
}
