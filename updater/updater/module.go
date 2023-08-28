package updater

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

func writeModule(dir string, module string, params []moduleParam, moduleConfig moduleConfig, mtime string) error {
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

func writeModuleFile(writer moduleWriter, file string, tmpfile string, params []moduleParam, moduleConfig moduleConfig) error {
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

		if !modified {
			newUID, newGID, newMode, err := moduleConfig.getNewFileAttrs(file)
			if err != nil {
				os.Remove(tmpfile)
				return err
			}
			if newUID != -1 || newGID != -1 || newMode != -1 {
				modified = true
			}
		}
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

	newUID, newGID, newMode, err := moduleConfig.getNewFileAttrs(tmpfile)
	if err != nil {
		os.Remove(tmpfile)
		return err
	}

	if newUID != -1 || newGID != -1 {
		err = os.Chown(tmpfile, newUID, newGID)
		if err != nil {
			os.Remove(tmpfile)
			return err
		}
	}

	if newMode != -1 {
		err = os.Chmod(tmpfile, os.FileMode(newMode))
		if err != nil {
			os.Remove(tmpfile)
			return err
		}
	}

	err = os.Rename(tmpfile, file)
	if err != nil {
		return err
	}

	log.Info().Str("file", file).Msg("file modified")
	return nil
}
