package updater

import (
	"bytes"
	"io/ioutil"

	"github.com/colinmarc/cdb"
)

type cdbWriter struct {
	file string
}

func newCdbWriter(file string) *cdbWriter {
	return &cdbWriter{file}
}

func (cw *cdbWriter) write(params []moduleParam) error {
	w, err := cdb.Create(cw.file)
	if err != nil {
		return err
	}

	for _, param := range params {
		var t string
		if param.json {
			t = "j"
		} else {
			t = "s"
		}
		err = w.Put([]byte(param.path), []byte(t+param.value))
		if err != nil {
			w.Close()
			return err
		}
	}

	return w.Close()
}

func (cw *cdbWriter) isModified(oldContent []byte) (bool, error) {
	newContent, err := ioutil.ReadFile(cw.file)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(oldContent, newContent), nil
}

func (cw *cdbWriter) close() error {
	return nil
}
