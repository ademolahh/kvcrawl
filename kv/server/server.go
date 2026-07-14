package server

import (
	"sync"

	"github.com/ademolahh/kvcrawl/kv/shared"
)

type KV struct {
	Data map[string]string
	mu   sync.RWMutex
}

func NewKV() *KV {
	return &KV{Data: make(map[string]string)}
}

func (kv *KV) Set(data *shared.SetArgs, result *bool) error {
	defer kv.mu.Unlock()
	kv.mu.Lock()
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
