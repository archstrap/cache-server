# Redis-Compatible Cache Server

A high-performance, Redis-compatible in-memory cache server implemented in Go. Built from first principles, this server implements the Redis Serialization Protocol (RESP), supports multiple data structures, master-replica replication, pub/sub messaging, ACID transactions, RDB persistence, and TTL-based expiry.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Features](#features)
- [Supported Commands](#supported-commands)
- [Data Structures](#data-structures)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Running the Server](#running-the-server)
- [Replication](#replication)
- [Persistence (RDB)](#persistence-rdb)
- [Pub/Sub](#pubsub)
- [Transactions](#transactions)
- [Protocol (RESP)](#protocol-resp)
- [Project Structure](#project-structure)
- [Concurrency Model](#concurrency-model)
- [TTL & Expiry](#ttl--expiry)

---

## Overview

This server is a ground-up implementation of a Redis-compatible cache, targeting wire-level compatibility with the Redis client ecosystem. It speaks RESP (REdis Serialization Protocol), enabling any standard Redis client (`redis-cli`, `go-redis`, `ioredis`, etc.) to connect and interact with it without modification.

The implementation covers the full spectrum from basic key-value storage through to advanced features including skip-list-based sorted sets, append-only streams, blocking list operations, multi-channel pub/sub, master-replica replication with offset tracking, and RDB snapshot loading.

---

## Architecture

```
                        ┌─────────────────────────────────────────────────────────┐
                        │                     TCP Server                          │
                        │              (internal/tcpserver)                       │
                        │   Accepts connections → spawns goroutine per client     │
                        └──────────────────────┬──────────────────────────────────┘
                                               │
                                               ▼
                        ┌─────────────────────────────────────────────────────────┐
                        │                    Event Loop                           │
                        │              (internal/eventloop)                       │
                        │   Buffered task channel → sequential per-connection     │
                        └──────────────────────┬──────────────────────────────────┘
                                               │
                             ┌─────────────────┼─────────────────┐
                             ▼                 ▼                 ▼
                    ┌─────────────┐   ┌─────────────┐   ┌──────────────┐
                    │   Parser    │   │  Transaction │   │  Replication │
                    │  (pkg/parser│   │    State     │   │   Propagation│
                    │  RESP I/O)  │   │  (shared/)   │   │  (internal/  │
                    └─────────────┘   └─────────────┘   │  replication)│
                                               │         └──────────────┘
                                               ▼
                        ┌─────────────────────────────────────────────────────────┐
                        │               Command Registry                          │
                        │             (internal/command)                          │
                        │  Normal Registry │ Special Registry (Pub/Sub, Ping)     │
                        └───────────────────────────────┬─────────────────────────┘
                                                        │
                  ┌─────────────────────────────────────┼──────────────────────────┐
                  ▼                   ▼                  ▼                          ▼
         ┌──────────────┐   ┌──────────────┐   ┌──────────────┐          ┌──────────────┐
         │  Cache Store │   │  List Store  │   │  Sorted Set  │          │ Stream Store │
         │  (String/TTL)│   │  (LPUSH etc) │   │  (Skip List) │          │  (XADD etc)  │
         └──────────────┘   └──────────────┘   └──────────────┘          └──────────────┘
```

### Design Principles

| Principle | Implementation |
|---|---|
| Wire compatibility | Full RESP protocol (arrays, bulk strings, integers, simple strings, errors) |
| Concurrency | One goroutine per connection; shared state protected by `sync.RWMutex` |
| Blocking operations | `sync.Cond` for BLPOP and XREAD BLOCK without busy-waiting |
| Command dispatch | Factory pattern with two registries (normal and special) |
| Expiry | Lazy evaluation — TTL checked on access, not via background sweeper |
| Replication | Handshake phase → RDB snapshot → command streaming |

---

## Features

- **RESP Protocol** — Full REdis Serialization Protocol v2 parser and serializer
- **Multiple Data Types** — Strings, Lists, Sorted Sets (Skip List), Streams
- **TTL / Expiry** — Per-key expiry via `EX` (seconds) and `PX` (milliseconds)
- **Transactions** — `MULTI` / `EXEC` / `DISCARD` with per-connection command queuing
- **Pub/Sub** — Multi-channel publish/subscribe with isolated subscriber context
- **Master-Replica Replication** — Full handshake, RDB snapshot delivery, write propagation, offset-based `WAIT`
- **RDB Persistence** — Load RDB v0011 snapshots on startup (keys, TTLs, metadata)
- **Blocking Operations** — `BLPOP` and `XREAD BLOCK` with configurable timeout
- **Pattern Matching** — Glob-style patterns in `KEYS` (e.g. `user:*:name`)
- **Config Flags** — CLI overrides for port, RDB path, and replica target

---

## Supported Commands

### String Commands

| Command | Syntax | Description |
|---|---|---|
| `SET` | `SET key value [EX seconds] [PX ms]` | Set a string value with optional TTL |
| `GET` | `GET key` | Get a string value; returns nil if expired or absent |
| `INCR` | `INCR key` | Atomically increment integer value by 1 |

### List Commands

| Command | Syntax | Description |
|---|---|---|
| `LPUSH` | `LPUSH key value [value ...]` | Prepend one or more values to a list |
| `RPUSH` | `RPUSH key value [value ...]` | Append one or more values to a list |
| `LPOP` | `LPOP key [count]` | Remove and return element(s) from the head |
| `BLPOP` | `BLPOP key [key ...] timeout` | Blocking pop with timeout (seconds) |
| `LRANGE` | `LRANGE key start stop` | Return a range of elements (0-indexed, negative indexing supported) |
| `LLEN` | `LLEN key` | Return the length of a list |

### Sorted Set Commands

| Command | Syntax | Description |
|---|---|---|
| `ZADD` | `ZADD key score member` | Add a member with a score |
| `ZRANGE` | `ZRANGE key start stop [WITHSCORES]` | Return members by rank range |
| `ZRANK` | `ZRANK key member` | Return zero-indexed rank of a member |
| `ZSCORE` | `ZSCORE key member` | Return the score of a member |
| `ZCARD` | `ZCARD key` | Return the number of members |
| `ZREM` | `ZREM key member [member ...]` | Remove one or more members |

### Stream Commands

| Command | Syntax | Description |
|---|---|---|
| `XADD` | `XADD key id field value [field value ...]` | Append an entry to a stream (auto ID with `*`) |
| `XRANGE` | `XRANGE key start end` | Return entries within an ID range |
| `XREAD` | `XREAD [COUNT n] [BLOCK ms] STREAMS key [key ...] id [id ...]` | Read entries from one or more streams |

### Transaction Commands

| Command | Syntax | Description |
|---|---|---|
| `MULTI` | `MULTI` | Start a transaction block |
| `EXEC` | `EXEC` | Execute all queued commands atomically |
| `DISCARD` | `DISCARD` | Abort the transaction and discard the queue |

### Pub/Sub Commands

| Command | Syntax | Description |
|---|---|---|
| `SUBSCRIBE` | `SUBSCRIBE channel [channel ...]` | Subscribe to one or more channels |
| `UNSUBSCRIBE` | `UNSUBSCRIBE [channel ...]` | Unsubscribe from channels |
| `PUBLISH` | `PUBLISH channel message` | Publish a message to a channel |
| `PING` | `PING [message]` | Heartbeat (works in pub/sub mode) |

### Replication Commands

| Command | Syntax | Description |
|---|---|---|
| `PSYNC` | `PSYNC replid offset` | Full resynchronization handshake |
| `REPLCONF` | `REPLCONF <subcommand> <value>` | Replication configuration and ACK exchange |
| `WAIT` | `WAIT numreplicas timeout` | Block until N replicas acknowledge the current offset |

### Key & Server Commands

| Command | Syntax | Description |
|---|---|---|
| `KEYS` | `KEYS pattern` | List all keys matching a glob pattern |
| `TYPE` | `TYPE key` | Return the data type stored at a key |
| `INFO` | `INFO [section]` | Return server information (replication section) |
| `CONFIG GET` | `CONFIG GET parameter` | Retrieve server configuration parameters |
| `ECHO` | `ECHO message` | Echo back the input message |
| `COMMAND` | `COMMAND` | Server acknowledgement |

---

## Data Structures

### String Store (`internal/store/cache.go`)

A thread-safe hash map keyed by string. Each entry is a `CacheItem` wrapping the raw value, an `expiresAt` timestamp (zero value = no expiry), and a `ValueType` discriminator.

```
CacheStore
└── map[string]*CacheItem
    ├── item       any        (string, []string, SkipList, StreamEntry, ...)
    ├── expiresAt  time.Time  (zero = immortal)
    └── valueType  ValueType  (string | list | zset | stream | hash)
```

### List Store (`internal/store/list_store.go`)

An ordered slice per key stored in a map of `List` containers. Supports both head (LPUSH/LPOP) and tail (RPUSH) operations. Blocking consumers wait on a `sync.Cond` that is broadcast whenever a new element is pushed.

### Sorted Set — Skip List (`internal/store/sorted_sets.go`)

Sorted sets are backed by a probabilistic skip list, providing O(log N) average complexity for insertions, deletions, rank queries, and range scans.

```
SkipList
├── head       *SkipNode            (sentinel; level[0] = sorted order)
├── index      map[string]*SkipNode (O(1) member lookup for ZSCORE / ZRANK)
├── level      int                  (current highest level, max 16)
└── size       int

SkipNode
├── member  string
├── score   float64
├── next    []*SkipNode  (forward pointers, one per level)
└── span    []int        (L0 nodes skipped per level, for O(log N) rank)
```

Level assignment uses a geometric distribution with P = 0.5 and a cap of 16 levels — identical to the Redis reference implementation.

### Stream Store (`internal/store/memory.go`)

Streams are append-only sequences of field-value maps, keyed by a monotonic `<milliseconds>-<sequence>` ID. A `sync.Cond` guards the backing slice so that `XREAD BLOCK` subscribers are woken precisely when new entries arrive.

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- `make` (optional, for convenience targets)

### Installation

```bash
git clone https://github.com/<org>/cache-server.git
cd cache-server
go mod download
```

### Configuration

**`resources/config.yaml`**

```yaml
port: 6379
host: "127.0.0.1"
maxParallelization: 1000
```

| Field | Default | Description |
|---|---|---|
| `port` | `6379` | TCP port to listen on |
| `host` | `127.0.0.1` | Bind address |
| `maxParallelization` | `1000` | Size of the internal task channel buffer |

**CLI Flags**

| Flag | Description |
|---|---|
| `--port <n>` | Override the port from config |
| `--dir <path>` | Directory containing the RDB file |
| `--dbfilename <name>` | RDB filename (requires `--dir`) |
| `--replicaof <host> <port>` | Start as a replica of the specified master |

### Running the Server

**Standalone (master)**

```bash
go run ./cmd/redis_server/
```

**With RDB snapshot loading**

```bash
go run ./cmd/redis_server/ --dir /var/lib/redis --dbfilename dump.rdb
```

**As a replica**

```bash
go run ./cmd/redis_server/ --port 6380 --replicaof 127.0.0.1 6379
```

**Connect with redis-cli**

```bash
redis-cli -p 6379 PING
redis-cli -p 6379 SET foo bar EX 60
redis-cli -p 6379 GET foo
```

---

## Replication

The server supports a single-level master-replica topology. A replica performs the full Redis replication handshake on startup and then receives a continuous stream of write commands from the master.

### Handshake Sequence

```
Replica                         Master
  │──── PING ──────────────────────▶│
  │◀─── +PONG ──────────────────────│
  │──── REPLCONF listening-port N ─▶│
  │◀─── +OK ────────────────────────│
  │──── REPLCONF capa psync2 ──────▶│
  │◀─── +OK ────────────────────────│
  │──── PSYNC ? -1 ────────────────▶│
  │◀─── +FULLRESYNC <replid> 0 ─────│
  │◀─── $<len>\r\n<rdb_bytes> ──────│  ← RDB snapshot
  │◀─── (write commands stream) ────│  ← ongoing propagation
```

### Write Propagation

Every write command processed by the master (SET, INCR, LPUSH, etc.) is serialized back to RESP and forwarded to all connected replicas. Offset tracking enables the `WAIT` command to block until at least N replicas have confirmed they have applied up to the current master offset.

### `INFO replication` Output

```
role:master
master_replid:8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb
master_repl_offset:0
connected_slaves:1
```

---

## Persistence (RDB)

On startup, if `--dir` and `--dbfilename` are both provided, the server parses the RDB file and restores all keys (including TTLs) into the cache before accepting connections.

### Supported RDB Opcodes

| Opcode | Hex | Description |
|---|---|---|
| `AUX` | `0xFA` | Auxiliary metadata (redis-ver, aof-base, etc.) |
| `RESIZEDB` | `0xFB` | Hash table and expire table size hints |
| `EXPIRETIME_MS` | `0xFC` | Key TTL in milliseconds (8-byte LE) |
| `EXPIRETIME` | `0xFD` | Key TTL in seconds (4-byte LE) |
| `SELECTDB` | `0xFE` | Database selector |
| `EOF` | `0xFF` | End of RDB file |

### String Encoding

| Prefix bits | Encoding |
|---|---|
| `00xxxxxx` | 6-bit length |
| `01xxxxxx xxxxxxxx` | 14-bit length |
| `10000000` + 4 bytes | 32-bit big-endian length |
| `11000000` | 8-bit integer |
| `11000001` | 16-bit integer |
| `11000010` | 32-bit integer |

---

## Pub/Sub

Clients enter pub/sub mode upon the first `SUBSCRIBE` command. In this mode only `PING`, `SUBSCRIBE`, `UNSUBSCRIBE`, and `QUIT` are accepted.

```bash
# Terminal 1 — subscriber
redis-cli SUBSCRIBE news alerts

# Terminal 2 — publisher
redis-cli PUBLISH news "breaking: server is up"
redis-cli PUBLISH alerts "disk usage at 90%"
```

Message format delivered to subscribers:

```
1) "message"
2) "news"
3) "breaking: server is up"
```

---

## Transactions

Commands issued between `MULTI` and `EXEC` are queued per-connection and executed atomically. The transaction is aborted and the queue discarded on `DISCARD`.

```bash
redis-cli MULTI
# OK
redis-cli SET foo 1
# QUEUED
redis-cli INCR foo
# QUEUED
redis-cli EXEC
# 1) OK
# 2) (integer) 2
```

Special commands (`SUBSCRIBE`, `PUBLISH`, `PING`) bypass the queue and execute immediately regardless of transaction state.

---

## Protocol (RESP)

All client-server communication uses RESP v2. Each message type is prefixed by a single byte:

| Prefix | Type | Example |
|---|---|---|
| `+` | Simple String | `+OK\r\n` |
| `-` | Error | `-ERR unknown command\r\n` |
| `:` | Integer | `:42\r\n` |
| `$` | Bulk String | `$5\r\nhello\r\n` |
| `*` | Array | `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n` |

A nil bulk string is represented as `$-1\r\n` and a nil array as `*-1\r\n`.

---

## Project Structure

```
.
├── cmd/
│   └── redis_server/
│       └── main.go                   # Entry point: signal handling, config, server start
├── internal/
│   ├── command/                      # One file per Redis command
│   │   ├── icommand.go               # ICommand interface
│   │   ├── scommand.go               # ISCommand interface (special registry)
│   │   ├── command_factory.go        # Registry construction and dispatch
│   │   ├── get_set.go                # GET, SET
│   │   ├── incr.go                   # INCR
│   │   ├── lpush.go / rpush.go       # LPUSH, RPUSH
│   │   ├── lpop.go / lpopblock.go    # LPOP, BLPOP
│   │   ├── lrange.go / llen.go       # LRANGE, LLEN
│   │   ├── zadd.go / zrange.go       # ZADD, ZRANGE
│   │   ├── zcard.go / zrank.go       # ZCARD, ZRANK
│   │   ├── zscore.go / zrem.go       # ZSCORE, ZREM
│   │   ├── xadd.go / xrange.go       # XADD, XRANGE
│   │   ├── xread.go                  # XREAD (with BLOCK support)
│   │   ├── multi.go / exec.go        # MULTI, EXEC
│   │   ├── discard.go                # DISCARD
│   │   ├── subscribe.go              # SUBSCRIBE
│   │   ├── unsubscribe.go            # UNSUBSCRIBE
│   │   ├── publish.go                # PUBLISH
│   │   ├── ping.go                   # PING
│   │   ├── keys.go                   # KEYS
│   │   ├── type.go                   # TYPE
│   │   ├── info.go                   # INFO
│   │   ├── psync.go                  # PSYNC
│   │   ├── repl_conf.go              # REPLCONF
│   │   ├── wait.go                   # WAIT
│   │   ├── config.go                 # CONFIG GET
│   │   ├── echo.go                   # ECHO
│   │   └── connect.go                # COMMAND
│   ├── config/
│   │   ├── config.go                 # YAML configuration loader (Viper)
│   │   └── flags.go                  # CLI flag definitions and parsing
│   ├── tcpserver/
│   │   └── server.go                 # TCP listener; spawns goroutine per connection
│   ├── eventloop/
│   │   ├── eventloop.go              # Buffered task channel consumer
│   │   └── redis_task.go             # Per-connection task: parse → dispatch → respond
│   ├── store/
│   │   ├── cache.go                  # String/generic KV store with TTL
│   │   ├── list_store.go             # List store with blocking support
│   │   ├── sorted_sets.go            # Skip list implementation for sorted sets
│   │   └── memory.go                 # Stream store with blocking XREAD
│   ├── replication/
│   │   ├── replication.go            # Master/Replica init, handshake, RDB delivery
│   │   └── store.go                  # Replica registry, write propagation, ACK tracking
│   ├── rdb/
│   │   └── reader.go                 # RDB v0011 parser: opcodes, string encoding, TTL
│   ├── shared/
│   │   ├── channels.go               # Pub/sub channel and subscriber store
│   │   ├── transaction.go            # Per-connection MULTI/EXEC/DISCARD state
│   │   ├── ack.go                    # REPLCONF ACK / WAIT acknowledgement state
│   │   └── command_processor.go      # CommandProcessor interface
│   └── scheduler/
│       └── scheduler.go              # (Reserved — background job scheduler)
├── pkg/
│   ├── model/
│   │   └── resp.go                   # RESP value types and command model
│   └── parser/
│       ├── paser.go                  # RESP input parser (reader → RespValue)
│       └── output_parser.go          # RESP output serializer (RespOutput → bytes)
├── util/
│   ├── byte.go                       # Byte-length utilities
│   ├── regex.go                      # Glob → regex conversion for KEYS
│   ├── gorouting_util.go             # OrDone: context-aware channel multiplexer
│   └── byte_test.go                  # Unit tests
├── test/
│   └── rdb.go                        # RDB test helpers
└── resources/
    └── config.yaml                   # Default server configuration
```

---

## Concurrency Model

The server uses a goroutine-per-connection model. All goroutines share a set of stores protected by `sync.RWMutex`. Blocking operations (BLPOP, XREAD BLOCK) use `sync.Cond` rather than polling to avoid wasting CPU while waiting for data.

| Component | Synchronization Primitive |
|---|---|
| Cache Store | `sync.RWMutex` (read-heavy optimised) |
| List Store | `sync.RWMutex` + `sync.Cond` (for BLPOP) |
| Stream Store | `sync.RWMutex` + `sync.Cond` (for XREAD BLOCK) |
| Replication Store | `sync.RWMutex` |
| Event Loop | Buffered `chan Task` (size = `maxParallelization`) |
| WAIT / ACK | Buffered `chan struct{}` per pending acknowledgement |

The event loop is intentionally kept as a simple channel consumer — there is no global lock on command execution. Isolation is achieved at the store level, allowing concurrent reads without contention.

---

## TTL & Expiry

Key expiry is implemented using a **lazy evaluation** strategy: TTL is not enforced by a background goroutine but is checked at the moment a key is accessed.

- `SET key value EX 30` — expires in 30 seconds
- `SET key value PX 5000` — expires in 5000 milliseconds
- `GET` returns nil and deletes the key if it has expired
- `KEYS` skips expired keys and removes them from the store during the scan

This matches Redis's default lazy expiry behaviour and avoids the overhead of a background sweeper for typical workloads.

---

## License

This project was built as part of the [CodeCrafters](https://codecrafters.io) "Build Your Own Redis" challenge.
