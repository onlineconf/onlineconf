package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/onlineconf/onlineconf/updater/v3/updater"
)

var configFile = flag.String("config", "/usr/local/etc/onlineconf.yaml", "config file")
var once = flag.Bool("once", false, "fetch configuration once and exit")

type AdminConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type ConfigFile struct {
	Hostname       string
	Datacenter     string
	Admin          AdminConfig
	DataDir        string `yaml:"data_dir"`
	UpdateInterval int    `yaml:"update_interval"`
	Variables      map[string]string
}

func main() {
	flag.Parse()
	config := readConfigFile(*configFile)
	u := updater.NewUpdater(*config)

	if *once {
		if u.Update() != nil {
			os.Exit(1)
		}
		return
	}

	log.Info().Msg("onlineconf-updater started")
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigC)
	go func() {
		sig := <-sigC
		log.Info().Str("signal", sig.String()).Msg("signal received, terminating")
		signal.Stop(sigC)
		u.Stop()
	}()

	u.Update()
	u.Run()
	log.Info().Msg("onlineconf-updater stopped")
}

func readConfigFile(filename string) *updater.UpdaterConfig {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to read config file")
	}
	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file")
	}

	var uri string
	switch config.Admin.Port {
	case 80:
		uri = "http://" + config.Admin.Host
	case 443, 0:
		uri = "https://" + config.Admin.Host
	default:
		uri = "http://" + config.Admin.Host + ":" + strconv.Itoa(config.Admin.Port)
	}

	return &updater.UpdaterConfig{
		Hostname:   config.Hostname,
		Datacenter: config.Datacenter,
		Admin: updater.AdminConfig{
			URI:      uri,
			Username: config.Admin.Username,
			Password: config.Admin.Password,
		},
		DataDir:        config.DataDir,
		UpdateInterval: time.Duration(config.UpdateInterval) * time.Second,
		Variables:      config.Variables,
	}
}
