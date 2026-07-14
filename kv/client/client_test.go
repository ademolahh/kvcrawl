package client

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"testing"

	"github.com/ademolahh/kvcrawl/kv/server"
)

func TestConcurrentClients(t *testing.T) {
	addr := newTestServer(t)
	const workers = 1000

	var wg sync.WaitGroup
	for i := range workers {

		wg.Go(func() {
			client, err := newTestClient(t, addr)
			if err != nil {
				t.Errorf("failed to connect to server: %v", err)
				return
			}
			defer client.client.Close()

			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)
			if err := client.set(key, value); err != nil {
				t.Errorf("set failed: %v", err)
				return
			}

			res, err := client.get(key)
			if err != nil {
				t.Errorf("get failed: %v", err)
				return
			}

			if value != res.Value {
				assertEq(t, value, res.Value)
			}
		})
	}

	wg.Wait()
}

func TestListCall(t *testing.T) {
	addr := newTestServer(t)
	client, err := newTestClient(t, addr)
	if err != nil {
		t.Errorf("failed to connect to server: %v", err)
		return
	}

	size := 10
	for i := range size {
		key := fmt.Sprintf("hello-%d", i)
		value := fmt.Sprintf("world-%d", i)
		if err := client.set(key, value); err != nil {
			t.Errorf("set failed: %v", err)
			return
		}
	}

	list, err := client.list()
	if err != nil {
		t.Errorf("list call failed: %v", err)
		return
	}

	if len(list) != size {
		assertEq(t, size, len(list))
	}

	if res, err := client.get(list[5]); err == nil && !res.Ok {
		assertEq(t, true, false)
	}
}

func TestDeleteCall(t *testing.T) {
	addr := newTestServer(t)
	client, err := newTestClient(t, addr)
	if err != nil {
		t.Errorf("failed to connect to server: %v", err)
		return
	}

	if err := client.set("hello", "world"); err != nil {
		t.Errorf("set failed: %v", err)
		return
	}

	ok, err := client.delete("hello")
	if err != nil {
		t.Errorf("error deleting key: %v", err)
		return
	}

	if !ok {
		assertEq(t, true, ok)
	}
}

func newTestClient(t *testing.T, addr net.Addr) (*Client, error) {
	t.Helper()

	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return &Client{
		client: rpc.NewClient(conn),
	}, nil
}

func newTestServer(t *testing.T) net.Addr {
	t.Helper()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	t.Cleanup(func() { listener.Close() })

	kv := server.NewKV()
	server := rpc.NewServer()
	server.Register(kv)

	go server.Accept(listener)

	return listener.Addr()
}

func assertEq(t *testing.T, expected, actual any) {
	t.Helper()
	t.Errorf("expected %v, actual %v", expected, actual)
}
