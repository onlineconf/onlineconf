package main

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

func writeModule(dir, module string, params []moduleParam, mtime string) error {
	wrapper := func(ext string, newWriter func(file string) moduleWriter) error {
		file := filepath.Join(dir, module+"."+ext)
		tmpfile := file + ".tmp"
		w := newWriter(tmpfile)
		err := writeModuleFile(w, file, tmpfile, params)
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

func writeModuleFile(writer moduleWriter, file, tmpfile string, params []moduleParam) error {
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
	if !modified {
		os.Remove(tmpfile)
		return nil
	}

	err = writer.close()
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
