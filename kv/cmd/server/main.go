package main

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/ademolahh/kvcrawl/kv/server"
)

func main() {
	kv := server.NewKV()
	rpc.Register(&kv)

	lst, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server listening on port 1234")

	rpc.Accept(lst)
}
