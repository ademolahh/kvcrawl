package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ademolahh/kvcrawl/kv/server"
)

func main() {
	kv := server.NewKV()
	rpc.Register(kv)

	lst, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	fmt.Println("Server listening on port 1234")

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		<-ch
		fmt.Println("Shutting down.....")

		if err := lst.Close(); err != nil {
			fmt.Printf("error closing listener: %v", err)
		}
	}()

	var wg sync.WaitGroup
	for {
		conn, err := lst.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Fatalf("accept error: %v", err)
			continue
		}

		wg.Add(1)
		go func(conn net.Conn) {
			defer wg.Done()
			defer conn.Close()

			rpc.ServeConn(conn)
		}(conn)

	}

	wg.Wait()

	fmt.Println("Server stopped")

}
