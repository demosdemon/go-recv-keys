package main

import (
	"bytes"
	"io"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

func serializeEntityList(el openpgp.EntityList, w io.Writer) error {
	var buf bytes.Buffer

	out, err := armor.Encode(&buf, openpgp.PublicKeyType, nil)
	if err != nil {
		return err
	}

	for _, e := range el {
		if err := e.Serialize(out); err != nil {
			return err
		}
	}

	if err := out.Close(); err != nil {
		return err
	}

	_, err = w.Write(buf.Bytes())
	return err
}
