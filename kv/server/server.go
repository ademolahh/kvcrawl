package main

import (
	"sync"

	"github.com/ademolahh/kvcrawl/kv/shared"
)

type KV struct {
	Data map[string]string
	mu   sync.RWMutex
}

func (kv *KV) Set(data *shared.SetArgs, result *bool) error {
	defer kv.mu.Unlock()
	defer kv.mu.Lock()
	kv.Data[data.Key] = data.Value
	*result = true
	return nil
}

func (kv *KV) Get(key string, value *string) error {
	defer kv.mu.RUnlock()
	kv.mu.RLock()
	*value = kv.Data[key]
	return nil
}
