package common

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var AdminConfig *ConfigFile = readConfigFile("/usr/local/etc/onlineconf-admin.yaml")

type DatabaseConfig struct {
	Host        string
	User        string
	Password    string
	Base        string
	Timeout     int
	MaxLifetime int `yaml:"max_lifetime"`
}

type ConfigFile struct {
	HTTP struct {
		Listen             string
		BehindReverseProxy bool `yaml:"behind_reverse_proxy"`
	}
	Auth struct {
		DatabaseConfig `yaml:",inline"`
		Table          string
		NameField      string `yaml:"name_field"`
		PasswordField  string `yaml:"password_field"`
		Condition      string
		Realm          string
	}
	Database             DatabaseConfig
	NotificationDatabase DatabaseConfig `yaml:"notification_database"`
}

func readConfigFile(filename string) *ConfigFile {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to read config file")
	}
	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file")
	}
	return &config
}
