package main

import (
	"fmt"
	"net"
	"net/rpc"
)

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
