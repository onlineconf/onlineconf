package main

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/onlineconf/onlineconf/updater/v3/updater"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type ClientCertificateConfig struct {
	Cert string
	Key  string
}

type AdminConfig struct {
	URI               string
	Host              string
	Port              int
	Username          string
	Password          string
	ClientCertificate ClientCertificateConfig `yaml:"client_certificate"`
}

type ConfigFile struct {
	Hostname       string
	Datacenter     string
	Admin          AdminConfig
	DataDir        string `yaml:"data_dir"`
	UpdateInterval int    `yaml:"update_interval"`
	Variables      map[string]string
}

func formURI(uri string) (validURI string, err error) {
	adminURL, err := url.Parse(uri)
	if err != nil {
		err = fmt.Errorf("url.Parse: %w", err)
		return
	}
	if adminURL.Scheme == "" || adminURL.Host == "" {
		err = fmt.Errorf("uri requires scheme and host")
		return
	}
	if len(adminURL.Path) > 0 && adminURL.Path[len(adminURL.Path)-1] == '/' {
		adminURL.Path = adminURL.Path[:len(adminURL.Path)-1]
	}
	validURI = adminURL.String()
	return
}

func deprecatedFormURI(host string, port int) (uri string) {
	switch port {
	case 80:
		uri = "http://" + host
	case 443, 0:
		uri = "https://" + host
	default:
		uri = "http://" + host + ":" + strconv.Itoa(port)
	}
	return uri
}

func parseClientCertificate(certCfg ClientCertificateConfig) ([]tls.Certificate, error) {
	if certCfg.Cert == "" || certCfg.Key == "" {
		if certCfg.Cert != "" || certCfg.Key != "" {
			return nil, fmt.Errorf("both key and cert fields need to be filled")
		}
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certCfg.Cert, certCfg.Key)
	if err != nil {
		return nil, fmt.Errorf("tls.LoadX509KeyPair: %w", err)
	}

	return []tls.Certificate{cert}, nil
}

func readConfigFile(filename string) *updater.UpdaterConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to read config file")
	}
	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file")
	}

	var uri string

	if config.Admin.URI != "" && (config.Admin.Host != "" || config.Admin.Port != 0) {
		log.Fatal().Str("file", filename).Msg("failed to parse config: both uri and host:port(deprecated) provided")
	}
	if config.Admin.URI != "" {
		uri, err = formURI(config.Admin.URI)
		if err != nil {
			log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file: uri invalid")
		}
	} else {
		uri = deprecatedFormURI(config.Admin.Host, config.Admin.Port)
	}

	certs, err := parseClientCertificate(config.Admin.ClientCertificate)
	if err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file: ClientCertificates invalid")
	}

	return &updater.UpdaterConfig{
		Hostname:   config.Hostname,
		Datacenter: config.Datacenter,
		Admin: updater.AdminConfig{
			URI:                uri,
			Username:           config.Admin.Username,
			Password:           config.Admin.Password,
			ClientCertificates: certs,
		},
		DataDir:        config.DataDir,
		UpdateInterval: time.Duration(config.UpdateInterval) * time.Second,
		Variables:      config.Variables,
	}
}
