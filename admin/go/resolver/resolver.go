package resolver

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

var treeI tree

func Initialize() {
	var ctx = context.Background()
	ctx = log.Logger.WithContext(ctx)
	if err := treeI.update(ctx); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to initialize tree")
	}
	go func() {
		c := time.Tick(5 * time.Second)
		for _ = range c {
			if err := treeI.update(ctx); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("failed to update tree")
			}
		}
	}()
}

func Synchronize(path string, target Synchronized) {
	ctx := log.Logger.WithContext(context.Background())
	treeI.synchronize(ctx, path, target)
}
