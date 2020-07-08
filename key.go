package main

import (
	"context"
	"time"

	"github.com/strothj/hkp"
)

type Key struct {
	input string
	hkp   *hkp.KeyID
}

func NewKey(key string) (*Key, error) {
	k, err := hkp.ParseKeyID(key)
	if err != nil {
		return nil, err
	}

	return &Key{key, k}, nil
}

func (k *Key) Resolve(ctx context.Context, timeout time.Duration, keyservers []*KeyServer) *Result {
	ch := make(chan *Result)
	defer close(ch)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, ks := range keyservers {
		go func(ks *KeyServer) {
			r := ks.Resolve(ctx, k)
			ch <- r
		}(ks)
	}

	var last *Result
	for idx := 0; idx < len(keyservers); idx++ {
		r := <-ch
		if last == nil {
			last = r
		}
		if last.Error != nil {
			last = r
		}
		if last.Error == nil {
			cancel()
		}
	}

	return last
}
