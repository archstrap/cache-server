# Pub/Sub Implementation: Your Go vs Redis C

This document compares your Go pub/sub implementation with the actual Redis source (`/Users/leo/projects/c/redis/src/pubsub.c`) and walks through how Redis solves the problem end-to-end.

---

## Table of Contents

1. [Your Implementation — How It Works](#your-implementation)
2. [Redis Implementation — How It Works](#redis-implementation)
3. [Side-by-Side Comparison](#side-by-side-comparison)
4. [Issues in Your Implementation](#issues-in-your-implementation)
5. [What Redis Does Better](#what-redis-does-better)
6. [Redis Flow — Deep Dive](#redis-flow-deep-dive)

---

## Your Implementation

### Data Structures

```
ChannelStore
├── details  map[net.Conn]*ChannelInfo   → per-client: which channels they joined
└── channels map[string][]net.Conn       → per-channel: which conns are subscribed

ChannelInfo
├── channels []string    → list of channel names
└── count    int         → (unused, always derived via len)
```

### Subscribe Flow

```
subscribeCommand
  └── channelStore.Subscribe(conn, channelName)
        ├── create ChannelInfo for conn if nil
        ├── append channelName to ChannelInfo.channels (slice)
        ├── append conn to channels[channelName] (slice)
        └── return ["subscribe", channelName, count]
```

### Unsubscribe Flow

```
unsubscribeCommand
  └── channelStore.Unsubscribe(conn, channelName)
        ├── ChannelInfo.Unsubscribe(channelName)   → linear scan + slice delete
        ├── linear scan of channels[channelName]   → remove conn
        ├── delete channel key if empty
        └── return ["unsubscribe", channelName, count]
```

### Publish Flow

```
publishCommand
  └── channelStore.Publish(channelName, message)
        ├── look up channels[channelName]
        ├── for each conn: build RESP array, write directly to socket
        └── return subscriber count
```

### Command Routing

Two registries exist:
- `HandlerFactory` — normal commands (GET, SET, …)
- `SRegistry` — pub/sub commands (SUBSCRIBE, UNSUBSCRIBE, PUBLISH, PING)

When a client is subscribed (`IsSubscribed(conn) == true`), only `SRegistry` commands are dispatched.

---

## Redis Implementation

### Data Structures

```c
// Per-client (client struct in server.h)
dict *pubsub_channels;       // channel name → NULL  (hash set)
dict *pubsub_patterns;       // pattern → NULL
dict *pubsubshard_channels;  // sharded channel → NULL

uint64_t flags;              // CLIENT_PUBSUB bit set when subscribed

// Per-server (redisServer struct in server.h)
kvstore *pubsub_channels;        // channel → dict{client*}  (slot-aware)
dict    *pubsub_patterns;        // pattern → dict{client*}
kvstore *pubsubshard_channels;   // sharded channel → dict{client*}

unsigned int pubsub_clients;     // total clients in pub/sub mode
```

Redis uses **two-level dictionaries**: both client-side (`client → channels`) and server-side (`channel → clients`) use hash maps, not slices.

### pubsubtype — Polymorphism in C

Redis unifies global and sharded pub/sub via a strategy struct:

```c
typedef struct pubsubtype {
    int    shard;
    dict  *(*clientPubSubChannels)(client*);   // fn ptr: client's channel dict
    int   (*subscriptionCount)(client*);        // fn ptr: count subscriptions
    kvstore **serverPubSubChannels;             // server-side channel→clients
    robj  **subscribeMsg;                       // shared "subscribe" bulk string
    robj  **unsubscribeMsg;                     // shared "unsubscribe" bulk string
    robj  **messageBulk;                        // shared "message" bulk string
} pubsubtype;

pubsubtype pubSubType;       // global pubsub
pubsubtype pubSubShardType;  // sharded pubsub (cluster-aware, slot-based)
```

Both subscribe and publish use the same code paths, just with a different `pubsubtype` instance.

### Subscribe Flow (pubsub.c:245)

```
subscribeCommand
  └── for each channel arg:
        pubsubSubscribeChannel(client, channel, pubSubType)
          ├── check if client->pubsub_channels already contains this channel (O(1))
          │     → return 0 (no-op) if already subscribed
          ├── dictAdd(client->pubsub_channels, channel, NULL)    O(1)
          ├── look up or create server.pubsub_channels[slot][channel]
          ├── dictAdd(server_channel_dict, client, NULL)          O(1)
          ├── incrRefCount(channel)                               refcount safety
          └── addReplyPubsubSubscribed(client, channel, totalCount)
  └── markClientAsPubSub(client)
        → client->flags |= CLIENT_PUBSUB
        → server.pubsub_clients++
```

### Unsubscribe Flow (pubsub.c:283)

```
unsubscribeCommand (no args = unsubscribe from ALL channels)
  └── for each channel:
        pubsubUnsubscribeChannel(client, channel, notify, pubSubType)
          ├── dictDelete(client->pubsub_channels, channel)  O(1)
          ├── find server.pubsub_channels[slot][channel]
          ├── dictDelete(server_channel_dict, client)       O(1)
          ├── if server_channel_dict is now empty → delete it (frees memory)
          ├── decrRefCount(channel)
          └── optionally: addReplyPubsubUnsubscribed(...)
  └── if clientTotalPubSubSubscriptionCount(c) == 0:
        unmarkClientAsPubSub(c)
          → client->flags &= ~CLIENT_PUBSUB
          → server.pubsub_clients--
```

### Publish Flow (pubsub.c:465)

```
publishCommand
  └── pubsubPublishMessageAndPropagateToCluster(channel, message)
        └── pubsubPublishMessageInternal(channel, message, pubSubType)
              ├── [DIRECT SUBSCRIBERS]
              │     look up server.pubsub_channels[slot][channel]
              │     for each client in that dict:
              │       addReplyPubsubMessage(client, channel, message)
              │       → push to client output buffer (NOT direct write)
              │
              ├── [PATTERN SUBSCRIBERS] (global pubsub only)
              │     iterate ALL server.pubsub_patterns
              │     for each pattern: stringmatchlen(pattern, channel)  glob match
              │     if match: addReplyPubsubPatMessage(client, pattern, channel, message)
              │
              └── return total receiver count
        └── forceCommandPropagation(AOF + replication)
        └── clusterPropagatePublish() if in cluster mode
```

### Message Format (RESP2)

```
# Direct channel message
*3\r\n
$7\r\nmessage\r\n
$<len>\r\n<channel>\r\n
$<len>\r\n<message>\r\n

# Pattern-matched message
*4\r\n
$8\r\npmessage\r\n
$<len>\r\n<pattern>\r\n
$<len>\r\n<channel>\r\n
$<len>\r\n<message>\r\n

# Subscribe confirmation
*3\r\n
$9\r\nsubscribe\r\n
$<len>\r\n<channel>\r\n
:<count>\r\n
```

---

## Side-by-Side Comparison

| Aspect | Your Go | Redis C |
|--------|---------|---------|
| **Client→channel lookup** | `[]string` (linear scan) | `dict` / hash map (O(1)) |
| **Channel→client lookup** | `[]net.Conn` (linear scan) | `dict{client*}` (O(1)) |
| **Duplicate subscribe check** | Missing — subscribes twice | `dictAdd` returns error on dup → no-op |
| **Unsubscribe with no args** | Not implemented | Unsubscribes from ALL channels |
| **Pattern subscribe (PSUBSCRIBE)** | Not implemented | Full glob matching |
| **Subscribed mode flag** | `map[net.Conn]` presence check | Bitflag `CLIENT_PUBSUB` on client struct |
| **Write to subscribers** | `conn.Write()` direct | `addReply` → buffered output queue |
| **Shared string objects** | New string built per message | Pre-allocated `robj*` shared across all msgs |
| **Cluster / sharded pubsub** | Not applicable (single node) | Full slot-aware kvstore + cluster propagation |
| **RESP3 support** | Not implemented | No restrictions in RESP3 subscribed mode |
| **AOF / replication** | Not applicable | `forceCommandPropagation()` on every PUBLISH |
| **Thread safety** | No mutex on ChannelStore | Global event loop (single-threaded I/O) |
| **Memory cleanup** | Empty channel slices removed | Empty server dicts deleted immediately |
| **Subscription count tracking** | `len(channels)` on each call | Separate functions per type (channels/patterns/shard) |

---

## Issues in Your Implementation

### 1. No duplicate-subscribe guard

```go
// channels.go:62-63
c.details[conn].Subscribe(channelName)
c.channels[channelName] = append(c.channels[channelName], conn)
```

If a client sends `SUBSCRIBE foo foo`, the same connection gets appended twice to `channels["foo"]`. On PUBLISH, that client receives the message twice. Redis's `dictAdd` silently rejects duplicates (returns `DICT_ERR`).

**Fix:** check before appending:
```go
func (c *ChannelStore) Subscribe(conn net.Conn, channelName string) []any {
    if info := c.details[conn]; info != nil {
        for _, ch := range info.channels {
            if ch == channelName {
                return []any{"subscribe", channelName, info.ChannelCount()}
            }
        }
    }
    // ... rest of subscribe
}
```

### 2. Slice-based lookups are O(N)

Both `ChannelInfo.channels []string` and `channels map[string][]net.Conn` require linear scans for unsubscribe and during `Remove()`. For large subscriber counts this is slow.

Redis uses hash maps (`dict`) for both directions — all lookups are O(1).

**Fix:** use `map[string]struct{}` for `ChannelInfo.channels` (set semantics) and `map[net.Conn]struct{}` for the per-channel subscriber list.

### 3. `UNSUBSCRIBE` with no arguments

Redis spec: `UNSUBSCRIBE` with no arguments unsubscribes from all channels. Your implementation reads `args[1]` directly and will panic or return a wrong response if the client sends bare `UNSUBSCRIBE`.

### 4. No mutex on `ChannelStore`

Your event loop spawns one goroutine per task (`go redisTask.exec()`), meaning concurrent goroutines share `cChannelStore` without synchronization. A concurrent `Subscribe` and `Remove` on different connections will cause a data race on the maps.

Redis sidesteps this entirely — it is single-threaded for I/O (one event loop, no concurrent map access). Your architecture needs a `sync.RWMutex` or `sync.Mutex` on `ChannelStore`.

### 5. `ChannelInfo.count` field is dead weight

```go
type ChannelInfo struct {
    channels []string
    count    int   // always out of sync; ChannelCount() uses len() anyway
}
```

`count` is set in `Subscribe` but never decremented in `Unsubscribe`. `ChannelCount()` correctly uses `len(c.channels)` — making `count` misleading and unused.

### 6. Direct `conn.Write()` in Publish

```go
// channels.go:127
conn.Write([]byte(result))
```

Redis never writes directly to the socket inside a command handler. It pushes to a per-client output buffer (`addReply`), and a separate I/O multiplexer (`ae` event loop) flushes buffers when the socket is writable. Direct writes can block the goroutine, fail silently on partial writes, and bypass any backpressure mechanism.

For a real implementation, push to a per-connection write channel and let a dedicated goroutine flush it.

---

## What Redis Does Better

### Buffered Output, Not Direct Writes

Every message delivery in Redis goes through `addReply()`:
- Appends to `client->buf` (static buffer) or `client->reply` (linked list of sds strings)
- The event loop's writable handler (`sendReplyToClient`) flushes when socket is ready
- Handles partial writes, backpressure, and client disconnect cleanly

### Shared String Objects

"message", "subscribe", "unsubscribe" are allocated once at startup as shared `robj*` objects. Every published message reuses these pointers — no string allocation per delivery.

### CLIENT_PUBSUB Flag vs Map Lookup

Redis tracks subscribed state as a **bitflag** on the client struct (`c->flags & CLIENT_PUBSUB`). Your implementation does a hash map lookup (`c.details[conn]`) on every single command dispatch. The bitflag approach is cache-friendly and zero-allocation.

### Clean Unsubscription Count

Redis tracks three independent subscription counts:
```c
clientSubscriptionsCount()       // global channels + patterns
clientShardSubscriptionsCount()  // shard channels
clientTotalPubSubSubscriptionCount() // all three
```
The `CLIENT_PUBSUB` flag is only cleared when **all** subscriptions (channels + patterns) reach zero.

---

## Redis Flow — Deep Dive

### Complete Subscribe Sequence

```
Client sends: *2\r\n$9\r\nSUBSCRIBE\r\n$3\r\nfoo\r\n

server.c: processCommand()
  └── check c->flags & CLIENT_PUBSUB
        → if set AND command not in whitelist → ERR reply
  └── call subscribeCommand(client)

pubsub.c: subscribeCommand()
  └── for each channel arg (c->argv[i]):
        pubsubSubscribeChannel(c, channel, pubSubType)
          ├── clientPubSubChannels(c) = c->pubsub_channels
          ├── dictAdd(c->pubsub_channels, channel, NULL)
          │     → returns DICT_ERR if already present → skip
          ├── slot = keyHashSlot(channel)
          ├── server_dict = kvstoreGetDict(server.pubsub_channels, slot)
          │     → create if not exists
          ├── dictAdd(server_dict, c, NULL)
          ├── incrRefCount(channel)
          └── addReplyPubsubSubscribed(c, channel, subscriptionCount(c))
                → addReplyArrayLen(c, 3)
                → addReply(c, shared.subscribebulk)   // "subscribe"
                → addReplyBulk(c, channel)
                → addReplyLongLong(c, count)
  └── markClientAsPubSub(c)
        → c->flags |= CLIENT_PUBSUB
        → server.pubsub_clients++

Event loop (ae.c): beforeSleep()
  └── handleClientsWithPendingWrites()
        └── for each client with pending output:
              writeToClient(c)
                └── write(c->fd, c->buf + c->sentlen, ...)
```

### Complete Publish Sequence

```
Client sends: *3\r\n$7\r\nPUBLISH\r\n$3\r\nfoo\r\n$5\r\nhello\r\n

server.c: processCommand() → publishCommand(client)

pubsub.c: publishCommand()
  └── receivers = pubsubPublishMessageAndPropagateToCluster(channel, message)
        └── pubsubPublishMessageInternal(channel, message, pubSubType)
              ├── slot = keyHashSlot(channel)
              ├── server_dict = kvstoreGetDict(server.pubsub_channels, slot)
              ├── if server_dict != NULL:
              │     de = dictGetIterator(server_dict)
              │     while (de = dictNext(iter)):
              │       receiver = dictGetKey(de)
              │       addReplyPubsubMessage(receiver, channel, message, shared.messagebulk)
              │         → addReplyArrayLen(receiver, 3)
              │         → addReply(receiver, shared.messagebulk)  // "message"
              │         → addReplyBulk(receiver, channel)
              │         → addReplyBulk(receiver, message)
              │
              ├── [pattern matching]
              │     di = dictGetIterator(server.pubsub_patterns)
              │     while (de = dictNext(di)):
              │       pattern = dictGetKey(de)
              │       if stringmatchlen(pattern, channel, 1):  // case-insensitive glob
              │         clients_dict = dictGetVal(de)
              │         for each client in clients_dict:
              │           addReplyPubsubPatMessage(client, pattern, channel, message)
              │
              └── return receivers
        └── forceCommandPropagation(c, PROPAGATE_AOF | PROPAGATE_REPL)
        └── if cluster: clusterPropagatePublish(channel, message, 0)
  └── addReplyLongLong(c, receivers)   // reply to publisher

Event loop: beforeSleep() → flush all pending writes to all receivers
```

### State Machine: Client Subscribed Mode

```
NORMAL MODE
    │
    │  SUBSCRIBE foo        (or PSUBSCRIBE, SSUBSCRIBE)
    ▼
SUBSCRIBED MODE  ←──────────────────────────────────────────────┐
    │                                                            │
    │  Allowed commands only:                                    │
    │    PING / SUBSCRIBE / PSUBSCRIBE / SSUBSCRIBE             │
    │    UNSUBSCRIBE / PUNSUBSCRIBE / SUNSUBSCRIBE               │
    │    QUIT / RESET                                            │
    │                                                            │
    │  UNSUBSCRIBE (last channel)                                │
    │  → clientTotalPubSubSubscriptionCount(c) == 0             │
    ▼                                                            │
NORMAL MODE                                                      │
    │                                                            │
    │  SUBSCRIBE again ───────────────────────────────────────── ┘
```

In RESP3, the subscribed mode restriction is lifted — clients can send any command interleaved with pub/sub.

---

## Summary: Where You Are Heading

Your implementation is **structurally correct** — the two-direction mapping (client→channels, channel→clients), the separate command registry for subscribed mode, and the connection cleanup on disconnect all mirror what Redis does.

The gaps are:
1. **O(N) data structures** where Redis uses O(1) hash maps
2. **Missing duplicate subscribe guard** — potential double-delivery
3. **No mutex** — data race under concurrent goroutines
4. **No `UNSUBSCRIBE` with zero args** — spec compliance issue
5. **Direct socket write in Publish** — bypasses backpressure, risks partial writes
6. **Dead `count` field** in ChannelInfo

Fix the mutex and duplicate guard first — those are correctness bugs. The slice→map migration is a performance improvement worth doing if you scale beyond a toy server.
