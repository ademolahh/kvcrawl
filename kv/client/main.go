package main

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/ademolahh/kvcrawl/kv/shared"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(conn)

	args := &shared.SetArgs{Key: "A", Value: "B"}
	var reply bool
	if err := client.Call("KV.Set", args, &reply); err != nil {
		panic(err)
	}

	fmt.Println("Response", reply)

	var value string
	if err := client.Call("KV.Get", "A", &value); err != nil {
		panic(err)
	}
	fmt.Println("Result is", value)
}
