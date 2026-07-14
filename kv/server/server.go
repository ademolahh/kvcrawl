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

func (kv *KV) Get(key shared.KeyArg, reply *shared.GetReply) error {
	defer kv.mu.RUnlock()
	kv.mu.RLock()
	reply.Value = kv.Data[key.Key]
	reply.Ok = true
	return nil
}

func (kv *KV) Delete(key shared.KeyArg, reply *bool) error {
	defer kv.mu.Unlock()
	kv.mu.Lock()

	if _, exists := kv.Data[key.Key]; exists {
		delete(kv.Data, key.Key)
		*reply = true
	}

	return nil

}

func (kv *KV) List(key shared.EmptyArgs, reply *shared.ListReply) error {
	defer kv.mu.RUnlock()
	kv.mu.RLock()

	for k := range kv.Data {
		reply.Values = append(reply.Values, kv.Data[k])
	}

	return nil
}
