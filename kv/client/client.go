package client

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

func (c *Client) get(key string) (*shared.GetReply, error) {
	args := shared.KeyArg{Key: key}
	var reply shared.GetReply
	if err := c.client.Call("KV.Get", args, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

func (c *Client) delete(key string) (bool, error) {
	args := shared.KeyArg{Key: key}
	var reply bool
	if err := c.client.Call("KV.Delete", args, &reply); err != nil {
		return false, nil
	}

	return true, nil
}

func (c *Client) list() ([]string, error) {
	var reply shared.ListReply
	if err := c.client.Call("KV.List", shared.EmptyArgs{}, &reply); err != nil {
		return nil, err
	}

	return reply.Values, nil
}

func (c *Client) Query(input string) (any, error) {
	msg := strings.Fields(strings.TrimSpace(input))
	if len(msg) == 0 {
		return nil, errors.New("invalid command")
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
	case "delete":
		if len(msg) != 2 {
			return "", errors.New("usage: delete <key>")
		}
		return c.delete(msg[1])
	case "list":
		return c.list()
	default:
		return "", errors.New("invalid command")
	}
}
