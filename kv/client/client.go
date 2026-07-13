package main

import (
	"errors"
	"net"
	"net/rpc"
	"strings"

	"github.com/ademolahh/kvcrawl/kv/shared"
)

type Client struct {
	client *rpc.Client
}

func NewClient(conn net.Conn) *Client {
	return &Client{client: rpc.NewClient(conn)}
}

func (c *Client) set(key, value string) error {
	args := &shared.SetArgs{Key: key, Value: value}
	var result bool
	return c.client.Call("KV.Set", args, &result)
}

func (c *Client) get(key string) (string, error) {
	var value string
	if err := c.client.Call("KV.Get", key, &value); err != nil {
		return "", err
	}

	return value, nil
}

func (c *Client) query(input string) (string, error) {
	msg := strings.Fields(strings.TrimSpace(input))
	if len(msg) == 0 {
		return "", errors.New("invalid command")
	}

	switch msg[0] {
	case "set":
		if len(msg) != 3 {
			return "", errors.New("usage: set <key> <value>")
		}
		err := c.set(msg[1], msg[2])
		return "", err
	case "get":
		if len(msg) != 2 {
			return "", errors.New("usage: get <key>")
		}
		return c.get(msg[1])
	default:
		return "", errors.New("invalid command")
	}
}
