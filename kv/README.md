# kv

A small in-memory key-value store served over TCP using Go's `net/rpc`, with a CLI client. State lives in a mutex-guarded map and is lost when the server stops.

## Layout

```
kv/
├── cmd/
│   ├── server/   # server entrypoint (listens on :1234)
│   └── client/   # client entrypoint (one command per invocation)
├── server/       # KV store: Set, Get, Delete, List RPC methods
├── client/       # RPC client and command parsing
└── shared/       # request/reply types shared by both sides
```

## Running

Start the server (listens on `localhost:1234`, shuts down gracefully on Ctrl-C):

```sh
go run ./kv/cmd/server
```

In another terminal, run commands with the client:

```sh
go run ./kv/cmd/client set name ada
go run ./kv/cmd/client get name
go run ./kv/cmd/client list
go run ./kv/cmd/client delete name
```

## Commands

| Command | Description |
| --- | --- |
| `set <key> <value>` | Store a value under a key (overwrites existing) |
| `get <key>` | Fetch the value for a key; errors if the key is missing |
| `delete <key>` | Remove a key |
| `list` | Return all stored values |

## Tests

```sh
go test -race ./kv/...
```
