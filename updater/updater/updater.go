package updater

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type AdminConfig struct {
	URI      string
	Username string
	Password string
}

type UpdaterConfig struct {
	Hostname       string
	Datacenter     string
	Admin          AdminConfig
	DataDir        string
	UpdateInterval time.Duration
	Variables      map[string]string
}

type Updater struct {
	config      UpdaterConfig
	adminClient *adminClient
	mtime       string
	done        chan struct{}
}

func NewUpdater(config UpdaterConfig) *Updater {
	if config.DataDir == "" {
		config.DataDir = "/usr/local/etc/onlineconf"
	}
	if config.UpdateInterval == time.Duration(0) {
		config.UpdateInterval = 10 * time.Second
	}
	return &Updater{
		config:      config,
		adminClient: newAdminClient(config.Admin),
		done:        make(chan struct{}),
	}
}

func (u *Updater) Run() {
	ticker := time.NewTicker(u.config.UpdateInterval)
	defer ticker.Stop()
	for {
		select {
		case <-u.done:
			return
		case <-ticker.C:
			u.Update()
		}
	}
}

func (u *Updater) Stop() {
	close(u.done)
}

func (u *Updater) Update() error {
	respMtime, modules, moduleConfigs, err := u.adminClient.getModules(u.config.Hostname, u.config.Datacenter, u.mtime, u.config.Variables)
	if err != nil {
		if err != ErrNotModified {
			log.Error().Err(err).Msg("failed to fetch config")
		}
		return err
	}
	log.Info().Strs("modules", modulesNames(modules)).Msg("")
	err = writeModules(u.config.DataDir, modules, moduleConfigs, respMtime)
	if err != nil {
		log.Error().Err(err).Msg("failed to write config files")
		return err
	}
	u.mtime = respMtime
	log.Info().Str("mtime", u.mtime).Msg("configuration updated")
	return nil
}

func modulesNames(m map[string][]moduleParam) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeModules(dir string, modules map[string][]moduleParam, moduleConfigs map[string]moduleConfig, mtime string) error {
	var err error
	for module, params := range modules {
		moduleConfig := moduleConfigs[module]
		err1 := writeModule(dir, module, params, moduleConfig, mtime)
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
