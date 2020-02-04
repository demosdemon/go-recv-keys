package main

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
)

type Result struct {
	Key       string
	KeyServer string
	Start     time.Time
	Duration  time.Duration
	Error     error
	Result    openpgp.EntityList
}

type serializedResult struct {
	Key       string        `json:"key" yaml:"key"`
	KeyServer string        `json:"keyserver" yaml:"keyserver"`
	Start     time.Time     `json:"start" yaml:"start"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
	Error     error         `json:"error,omitempty" yaml:"error,omitempty"`
	Result    string        `json:"result,omitempty" yaml:"result,omitempty"`
}

func (r *Result) serialize() (*serializedResult, error) {
	var w bytes.Buffer
	if len(r.Result) > 0 {
		if err := serializeEntityList(r.Result, &w); err != nil {
			return nil, errors.Wrap(err, "unable to serialize entity list")
		}
		_, _ = w.WriteString("\n")
	}
	return &serializedResult{
		Key:       r.Key,
		KeyServer: r.KeyServer,
		Start:     r.Start,
		Duration:  r.Duration,
		Error:     r.Error,
		Result:    w.String(),
	}, nil
}

func (r *Result) deserialize(res *serializedResult) error {
	if len(res.Result) > 0 {
		reader := bytes.NewBufferString(res.Result)
		el, err := openpgp.ReadArmoredKeyRing(reader)
		if err != nil {
			return err
		}
		r.Result = el
	} else {
		r.Result = nil
	}

	r.Key = res.Key
	r.KeyServer = res.KeyServer
	r.Start = res.Start
	r.Duration = res.Duration
	r.Error = res.Error
	return nil
}

func (r *Result) MarshalJSON() ([]byte, error) {
	s, err := r.serialize()
	if err != nil {
		return nil, err
	}
	return json.Marshal(s)
}

func (r *Result) UnmarshalJSON(data []byte) error {
	var s serializedResult
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	return r.deserialize(&s)
}

func (r *Result) MarshalYAML() (interface{}, error) {
	return r.serialize()
}

func (r *Result) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s serializedResult
	if err := unmarshal(&s); err != nil {
		return err
	}
	return r.deserialize(&s)
}
