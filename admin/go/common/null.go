package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"reflect"

	"github.com/ugorji/go/codec"
)

type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	} else {
		return json.Marshal(nil)
	}
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.String)
}

func (ns *NullString) CodecEncodeSelf(enc *codec.Encoder) {
	if ns.Valid {
		enc.MustEncode(ns.String)
	} else {
		enc.MustEncode(nil)
	}
}

func (ns *NullString) CodecDecodeSelf(dec *codec.Decoder) {
	var v interface{}
	dec.MustDecode(&v)
	value := reflect.ValueOf(v)
	if !value.IsNil() {
		ns.Valid = true
		ns.String = value.String()
	}
}

type NullBool struct {
	sql.NullBool
}

func (ns NullBool) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.Bool)
	} else {
		return json.Marshal(nil)
	}
}

func (ns *NullBool) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.Bool)
}
