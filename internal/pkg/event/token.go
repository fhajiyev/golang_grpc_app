package event

import (
	"bytes"
	"encoding/gob"

	"github.com/Buzzvil/buzzlib-go/jwe"
)

// TokenEncrypter interface definition
type TokenEncrypter interface {
	Build(token Token) (string, error)
	Parse(tokenStr string) (*Token, error)
}

type tokenEncrypter struct {
	manager jwe.Manager
}

func (t *tokenEncrypter) Build(token Token) (string, error) {
	serialized, err := t.serialize(token)
	if err != nil {
		return "", err
	}

	return t.manager.GenerateToken(serialized)
}

func (t *tokenEncrypter) Parse(tokenStr string) (*Token, error) {
	serialized, err := t.manager.GetDataByToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return t.deserialize(serialized)
}

func (t *tokenEncrypter) serialize(token Token) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)

	err := e.Encode(token)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (t *tokenEncrypter) deserialize(serialized []byte) (*Token, error) {
	b := bytes.Buffer{}
	b.Write(serialized)
	d := gob.NewDecoder(&b)

	token := Token{}
	err := d.Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
