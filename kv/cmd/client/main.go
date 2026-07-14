package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ademolahh/kvcrawl/kv/client"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := client.NewClient(conn)

	input := strings.Join(os.Args[1:], " ")

	data, err := client.Query(input)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if data != "" {
		fmt.Println("Result is", data)
	}

}
