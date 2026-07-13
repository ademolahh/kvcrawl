package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := NewClient(conn)
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		data, err := client.query(input)
		if err != nil {
			continue
		}
		if data != "" {
			fmt.Println("Result is", data)
		}
	}
}
