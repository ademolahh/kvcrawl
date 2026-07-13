package main

import (
	"fmt"
	"net"
	"net/rpc"
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

func main() {
	kv := KV{Data: make(map[string]string)}
	rpc.Register(&kv)

	lst, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server listening on port 1234")

	rpc.Accept(lst)
}
