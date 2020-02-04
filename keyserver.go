package main

import (
	"context"
	"time"

	"github.com/strothj/hkp"
)

type KeyServer struct {
	input string
	hkp   *hkp.Keyserver
}

func NewKeyServer(keyserver string) (*KeyServer, error) {
	ks, err := hkp.ParseKeyserver(keyserver)
	if err != nil {
		return nil, err
	}

	return &KeyServer{keyserver, ks}, nil
}

func (ks *KeyServer) Resolve(ctx context.Context, k *Key) *Result {
	client := hkp.NewClient(ks.hkp, httpClient)
	start := time.Now()
	el, err := client.GetKeysByID(ctx, k.hkp)
	duration := time.Since(start)
	return &Result{
		Key:       k.input,
		KeyServer: ks.input,
		Start:     start,
		Duration:  duration,
		Error:     err,
		Result:    el,
	}
}
