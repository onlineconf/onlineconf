package updater

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

var ErrNotModified = errors.New("config not modified")

var templateRe = regexp.MustCompile(`\$\{(.*?)\}`)

type adminClient struct {
	client *http.Client
	config AdminConfig
}

func newAdminClient(config AdminConfig) *adminClient {
	client := &http.Client{}
	return &adminClient{
		config: config,
		client: client,
	}
}

func (a *adminClient) getModules(hostname, datacenter, mtime string, vars map[string]string) (string, map[string][]moduleParam, map[string]moduleConfig, error) {
	respMtime, data, err := a.getConfigData(hostname, datacenter, mtime)
	if err != nil {
		return "", nil, nil, err
	}
	modules, moduleConfigs := prepareModules(data, vars)
	return respMtime, modules, moduleConfigs, nil
}

func (a *adminClient) getConfigData(hostname, datacenter, mtime string) (string, *ConfigData, error) {
	req, err := http.NewRequest("GET", a.config.URI+"/client/config", nil)
	if err != nil {
		return "", nil, err
	}

	if hostname != "" {
		req.Header.Add("X-OnlineConf-Client-Host", hostname)
	}
	if datacenter != "" {
		req.Header.Add("X-OnlineConf-Client-Datacenter", datacenter)
	}
	req.Header.Add("X-OnlineConf-Client-Version", version)
	req.SetBasicAuth(a.config.Username, a.config.Password)

	req.Header.Add("X-OnlineConf-Client-Mtime", mtime)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		respMtime := resp.Header.Get("X-OnlineConf-Admin-Last-Modified")
		data, err := deserializeConfigData(resp.Body)
		return respMtime, data, err
	case 304:
		return "", nil, ErrNotModified
	default:
		return "", nil, errors.New(resp.Status)
	}
}

func prepareModules(data *ConfigData, vars map[string]string) (modules map[string][]moduleParam, moduleConfigs map[string]moduleConfig) {
	moduleConfigs = make(map[string]moduleConfig, len(data.Modules))
	modules = make(map[string][]moduleParam, len(data.Modules))
	for _, m := range data.Modules {
		modules[m] = []moduleParam{}
	}
	sort.Slice(data.Nodes, func(i, j int) bool {
		return data.Nodes[i].Path < data.Nodes[j].Path
	})
	defaultModuleConfig := moduleConfig{}
	delimiter := defaultModuleConfig.Delimiter
	for _, param := range data.Nodes {
		if !strings.HasPrefix(param.Path, "/onlineconf/module/") {
			switch param.Path {
			case "/", "/onlineconf":
			case "/onlineconf/module":
				defaultModuleConfig = readModuleConfig(param)
			default:
				log.Warn().Str("path", param.Path).Msg("parameter is out of '/onlineconf/module/' subtree")
			}
			continue
		}

		pc := strings.Split(param.Path, "/")
		moduleName := pc[3]
		if modules[moduleName] == nil {
			modules[moduleName] = []moduleParam{}
		}
		if _, ok := moduleConfigs[moduleName]; !ok {
			moduleConfigs[moduleName] = defaultModuleConfig
		}

		if len(pc) == 4 {
			moduleConfig := readModuleConfig(param)
			delimiter = moduleConfig.Delimiter
			moduleConfigs[moduleName] = moduleConfig
			if delimiter == "" {
				if v := defaultModuleConfig.Delimiter; v != "" {
					delimiter = v
				} else if moduleName == "TREE" {
					delimiter = "/"
				} else {
					delimiter = "."
				}
			}
			continue
		}
		if param.ContentType == "application/x-null" {
			continue
		}

		var mParam moduleParam
		if delimiter == "/" {
			mParam.path = strings.Join(append([]string{""}, pc[4:]...), "/")
		} else {
			mParam.path = strings.Join(pc[4:], ".")
		}

		switch param.ContentType {
		case "application/json":
			var buf bytes.Buffer
			err := json.Compact(&buf, []byte(param.Value.String))
			if err != nil {
				log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to compact json")
				continue
			}
			mParam.value = buf.String()
			mParam.json = true
			if mParam.value == "null" {
				continue
			}
		case "application/x-yaml":
			v, err := YAMLToJSON([]byte(param.Value.String))
			if err != nil {
				log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to convert yaml to json")
				continue
			}
			mParam.value = string(v)
			mParam.json = true
			if mParam.value == "null" {
				continue
			}
		case "application/x-template":
			mParam.value = templateRe.ReplaceAllStringFunc(param.Value.String, func(match string) string {
				name := match[2 : len(match)-1]
				return vars[name]
			})
		default:
			mParam.value = param.Value.String
		}

		modules[moduleName] = append(modules[moduleName], mParam)
	}
	return
}

type moduleConfig struct {
	Delimiter string
	Owner     string
	Mode      string
	usr       *user.User
	grp       *user.Group
}

type ModuleConfigStatus int

const (
	FileNotChanged ModuleConfigStatus = iota
	FileModeChanged
	FileOwnerChanged
)

func (config moduleConfig) getFileInfo(file string) (int, int, *syscall.Stat_t, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0, 0, nil, err
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, nil, fmt.Errorf("can't fetch stat info for file %s", file)
	}

	return int(stat.Uid), int(stat.Gid), stat, nil
}

func (config *moduleConfig) getStatus(file string) (ModuleConfigStatus, error) {
	if config.Owner == "" && config.Mode == "" {
		return FileNotChanged, nil
	}

	newUsr, newGrp, err := config.parseOwnerString()
	if err != nil {
		return FileNotChanged, err
	}

	fileUID, fileGID, fileInfo, err := config.getFileInfo(file)
	if err != nil {
		if os.IsNotExist(err) {
			return FileNotChanged, nil
		}
		return FileNotChanged, err
	}

	fileUidStr := strconv.Itoa(fileUID)
	fileGidStr := strconv.Itoa(fileGID)

	status := FileNotChanged
	if fileUidStr != newUsr.Uid || fileGidStr != newGrp.Gid {
		status |= FileOwnerChanged
	}

	if mode := config.Mode; mode != "" {
		modeUint, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return FileNotChanged, fmt.Errorf("convert file mode %s failure...%v", mode, err)
		}

		if os.FileMode(modeUint) != os.FileMode(fileInfo.Mode) {
			status |= FileModeChanged
		}
	}

	return status, nil
}

func (config *moduleConfig) parseOwnerString() (*user.User, *user.Group, error) {

	// return object if already init
	if config.usr != nil && config.grp != nil {

		return config.usr, config.grp, nil
	}

	list := strings.Split(config.Owner, ":")
	if len(list) != 2 {
		return nil, nil, fmt.Errorf("wrong owner format: '%s'", config.Owner)
	}

	var err error
	usrStr, grpStr := list[0], list[1]

	// init user object
	var usr *user.User
	if _, err = strconv.ParseInt(usrStr, 10, 64); err == nil {

		usr, err = user.LookupId(usrStr)
		if err != nil {
			return nil, nil, err
		}
	} else {

		usr, err = user.Lookup(usrStr)
		if err != nil {
			return nil, nil, err
		}
	}

	// init group object
	var grp *user.Group
	if _, err = strconv.ParseInt(grpStr, 10, 64); err == nil {

		grp, err = user.LookupGroupId(grpStr)
		if err != nil {
			return nil, nil, err
		}
	} else {

		grp, err = user.LookupGroup(grpStr)
		if err != nil {
			return nil, nil, err
		}
	}

	config.usr = usr
	config.grp = grp

	return usr, grp, nil
}

func (config *moduleConfig) changeOwner(file string) error {
	newUsr, newGrp, err := config.parseOwnerString()
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(newUsr.Uid)
	if err != nil {
		return fmt.Errorf("convert owner %s uid %s failure...%v", config.Owner, newUsr.Uid, err)
	}

	gid, err := strconv.Atoi(newGrp.Gid)
	if err != nil {
		return fmt.Errorf("convert owner %s gid %s failure...%v", config.Owner, newGrp.Gid, err)
	}

	err = os.Chown(file, uid, gid)
	if err != nil {
		return fmt.Errorf("chown file %s failure...%v", file, err)
	}

	return nil
}

func (config *moduleConfig) changeMode(file string, tmpfile string) error {
	modeUint, err := strconv.ParseUint(config.Mode, 8, 32)
	if err != nil {
		return fmt.Errorf("convert file mode %s failure...%v", config.Mode, err)
	}

	fileUID, fileGID, _, err := config.getFileInfo(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		fileUID = os.Getuid()
		fileGID = os.Getgid()
	}

	err = os.Chmod(tmpfile, os.FileMode(modeUint))
	if err != nil {
		return fmt.Errorf("chmod file %s failure...%v", tmpfile, err)
	}

	err = os.Chown(tmpfile, fileUID, fileGID)
	if err != nil {
		return fmt.Errorf("chown file %s failure...%v", tmpfile, err)
	}

	return nil
}

func readModuleConfig(param ConfigParam) (cfg moduleConfig) {
	var err error
	value := []byte(param.Value.String)
	switch param.ContentType {
	case "application/x-yaml":
		value, err = YAMLToJSON(value)
		if err != nil {
			log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to convert yaml to json")
			break
		}
		fallthrough
	case "application/json":
		err = json.Unmarshal(value, &cfg)
		if err != nil {
			log.Warn().Err(err).Str("param", param.Path).Str("value", string(value)).Msg("failed to parse json")
		}
	}
	return
}
