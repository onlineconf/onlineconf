package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

const version = "VERSION"

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
	Admin          AdminConfig
	DataDir        string `yaml:"data_dir"`
	UpdateInterval int    `yaml:"update_interval"`
}

func main() {
	flag.Parse()
	config := readConfigFile(*configFile)

	if *once {
		var mtime string
		if update(config, &mtime) != nil {
			os.Exit(1)
		}
		return
	}

	log.Info().Msg("onlineconf-updater started")
	done := make(chan struct{})
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigC)
	go func() {
		sig := <-sigC
		log.Info().Str("signal", sig.String()).Msg("signal received, terminating")
		signal.Stop(sigC)
		close(done)
	}()

	var mtime string
	update(config, &mtime)

	ticker := time.NewTicker(time.Duration(config.UpdateInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			log.Info().Msg("onlineconf-updater stopped")
			return
		case <-ticker.C:
			update(config, &mtime)
		}
	}
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

func update(config *ConfigFile, mtime *string) error {
	respMtime, modules, err := getModules(config.Admin, config.Hostname, *mtime)
	if err != nil {
		if err != ErrNotModified {
			log.Error().Err(err).Msg("failed to fetch config")
		}
		return err
	}
	err = writeModules(config.DataDir, modules, respMtime)
	if err != nil {
		log.Error().Err(err).Msg("failed to write config files")
		return err
	}
	*mtime = respMtime
	log.Info().Str("mtime", *mtime).Msg("configuration updated")
	return nil
}

func writeModules(dir string, modules map[string][]moduleParam, mtime string) error {
	var err error
	for module, params := range modules {
		err1 := writeModule(dir, module, params, mtime)
		if err1 != nil && err == nil {
			err = err1
		}
	}
	if err != nil {
		return err
	}
	return removeOldFiles(dir, modules)
}

func removeOldFiles(dir string, modules map[string][]moduleParam) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(0)
	if err != nil {
		return err
	}

	for _, name := range names {
		var module string
		if strings.HasSuffix(name, ".cdb") {
			module = strings.TrimSuffix(name, ".cdb")
		} else if strings.HasSuffix(name, ".conf") {
			module = strings.TrimSuffix(name, ".conf")
		}
		if module == "" {
			continue
		}
		_, ok := modules[module]
		if ok {
			continue
		}

		file := filepath.Join(dir, name)
		err1 := os.Remove(file)
		if err1 != nil {
			log.Error().Err(err1).Str("file", file).Msg("failed to remove file")
			if err == nil {
				err = err1
			}
			continue
		}
		log.Info().Str("module", module).Str("file", file).Msg("removed old module file")
	}

	return err
}
