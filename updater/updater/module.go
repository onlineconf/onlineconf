package updater

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

type moduleParam struct {
	path  string
	json  bool
	value string
}

type moduleWriter interface {
	write(params []moduleParam) error
	isModified(oldContent []byte) (bool, error)
	close() error
}

func writeModule(
	dir string,
	module string,
	params []moduleParam,
	moduleConfig *moduleConfig,
	mtime string,
) error {
	wrapper := func(ext string, newWriter func(file string) moduleWriter) error {
		file := filepath.Join(dir, module+"."+ext)
		tmpfile := file + ".tmp"
		w := newWriter(tmpfile)
		err := writeModuleFile(w, file, tmpfile, params, moduleConfig)
		if err != nil {
			log.Error().Err(err).Str("module", module).Msgf("failed to write .%s file", ext)
		}
		return err
	}
	cdbErr := wrapper("cdb", func(file string) moduleWriter {
		return newCdbWriter(file)
	})
	confErr := wrapper("conf", func(file string) moduleWriter {
		return newConfWriter(file, module, mtime)
	})
	if cdbErr != nil {
		return cdbErr
	}
	return confErr
}

func writeModuleFile(
	writer moduleWriter,
	file string,
	tmpfile string,
	params []moduleParam,
	moduleConfig *moduleConfig,
) error {
	err := writer.write(params)
	if err != nil {
		os.Remove(tmpfile)
		return err
	}

	modified := false
	oldContent, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			modified = true
		} else {
			os.Remove(tmpfile)
			return err
		}
	}
	if !modified {
		modified, err = writer.isModified(oldContent)
		if err != nil {
			os.Remove(tmpfile)
			return err
		}
	}

	// owner or mode were set
	if !moduleConfig.Empty() {
		modified = true
	}

	if !modified {
		os.Remove(tmpfile)
		return nil
	}

	err = writer.close()
	if err != nil {
		os.Remove(tmpfile)
		return err
	}

	err = chownModuleFile(tmpfile, moduleConfig)
	if err != nil {
		os.Remove(tmpfile)
		return err
	}

	err = os.Rename(tmpfile, file)
	if err != nil {
		return err
	}

	log.Info().Str("file", file).Msg("file modified")
	return nil
}

// change file owner and mode
func chownModuleFile(file string, moduleConfig *moduleConfig) error {
	if owner := moduleConfig.Owner; owner != "" {
		user, err := user.Lookup(owner)
		if err != nil {
			return fmt.Errorf("get owner %s info failure...%v", owner, err)
		}

		uid, err := strconv.Atoi(user.Uid)
		if err != nil {
			return fmt.Errorf("convert owner %s uid %s failure...%v", owner, user.Uid, err)
		}

		gid, err := strconv.Atoi(user.Gid)
		if err != nil {
			return fmt.Errorf("convert owner %s gid %s failure...%v", owner, user.Gid, err)
		}

		err = os.Chown(file, uid, gid)
		if err != nil {
			return fmt.Errorf("chown file %s failure...%v", file, err)
		}
	}
	if mode := moduleConfig.Mode; mode != "" {
		modeUint, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return fmt.Errorf("convert file mode %s failure...%v", mode, err)
		}

		err = os.Chmod(file, os.FileMode(modeUint))
		if err != nil {
			return fmt.Errorf("chmod file %s failure...%v", file, err)
		}
	}
	return nil
}
