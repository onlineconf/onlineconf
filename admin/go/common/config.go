package common

import (
	"os"

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

// EncryptionKeyFileEnv names the key file when the configuration file does
// not (the Docker image maps it to encryption.key_file as well).
const EncryptionKeyFileEnv = "ONLINECONF_ENCRYPTION_KEY_FILE"

// DefaultEncryptionKeyFile is where the sample configuration and the Docker
// image put the key file. It is never loaded implicitly: a file found there
// without a configuration referencing it is reported, not used.
const DefaultEncryptionKeyFile = "/usr/local/etc/onlineconf-admin.keys"

func CommonInitialize(c CommonConfig) {
	DB = OpenDatabase(c.Database)
	initializeKeyring(c.Encryption)
}

// initializeKeyring loads the key file named by the configuration or the
// environment, and always says in the log whether value encryption is on.
func initializeKeyring(c EncryptionConfig) {
	keyFile, source := c.KeyFile, "encryption.key_file"
	if keyFile == "" {
		if env := os.Getenv(EncryptionKeyFileEnv); env != "" {
			keyFile, source = env, EncryptionKeyFileEnv
		}
	}
	if keyFile == "" {
		if _, err := os.Stat(DefaultEncryptionKeyFile); err == nil {
			log.Warn().Str("file", DefaultEncryptionKeyFile).
				Msgf("value encryption is not configured: a key file exists at the default path but neither encryption.key_file nor %s names it", EncryptionKeyFileEnv)
		} else {
			log.Info().Msgf("value encryption is not configured: set encryption.key_file (or %s) to enable it", EncryptionKeyFileEnv)
		}
		return
	}
	keys, err := LoadKeyring(keyFile)
	if err != nil {
		log.Fatal().Err(err).Str("source", source).Msg("failed to load the encryption key file")
	}
	Keys = keys
	ids := make([]string, 0, len(keys.All()))
	for _, key := range keys.All() {
		ids = append(ids, key.ID)
	}
	log.Info().Str("file", keyFile).Str("source", source).Str("active", keys.Active().ID).Strs("keys", ids).Msg("encryption keyring loaded")
}
