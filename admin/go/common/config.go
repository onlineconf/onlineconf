package common

import (
	"github.com/rs/zerolog/log"
)

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	Base        string
	Timeout     int
	MaxLifetime int `yaml:"max_lifetime"`
	MaxConn     int `yaml:"max_conn"`
}

type CommonConfig struct {
	Database   DatabaseConfig
	Encryption EncryptionConfig
}

func CommonInitialize(c CommonConfig) {
	DB = OpenDatabase(c.Database)
	initializeKeyring(c.Encryption)
}

// initializeKeyring loads the key file named by the configuration, and
// always says in the log whether value encryption is on.
func initializeKeyring(c EncryptionConfig) {
	if c.KeyFile == "" {
		log.Info().Msg("value encryption is not configured: set encryption.key_file to enable it")
		return
	}
	keys, err := LoadKeyring(c.KeyFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load the encryption key file")
	}
	Keys = keys
	ids := make([]string, 0, len(keys.All()))
	for _, key := range keys.All() {
		ids = append(ids, key.ID)
	}
	log.Info().Str("file", c.KeyFile).Strs("keys", ids).Msg("encryption keyring loaded")
}
