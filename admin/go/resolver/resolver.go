package resolver

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var treeI tree

func Initialize() {
	var ctx = context.Background()
	ctx = log.Logger.WithContext(ctx)
	if err := treeI.update(ctx); err != nil {
		if IsDecryptionError(err) {
			// the keyring cannot open a stored value: serving would mean
			// serving ciphertext or dropping secrets — refuse to start
			log.Ctx(ctx).Fatal().Err(err).Msg("failed to initialize tree: an encrypted value cannot be decrypted with the configured keys")
		}
		log.Ctx(ctx).Error().Err(err).Msg("failed to initialize tree")
	}
	go func() {
		c := time.Tick(5 * time.Second)
		for range c {
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
