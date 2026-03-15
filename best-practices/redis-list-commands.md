# Redis List Commands — Best Practices

Learned from Redis source code at `/Users/leo/projects/c/redis/src/`.

---

## Source File Map

| Command | Source File | Handler Function | Generic/Helper |
|---------|------------|-----------------|----------------|
| LPUSH   | `src/t_list.c:508` | `lpushCommand()` | `pushGenericCommand(c, LIST_HEAD, 0)` |
| RPUSH   | `src/t_list.c:513` | `rpushCommand()` | `pushGenericCommand(c, LIST_TAIL, 0)` |
| LLEN    | `src/t_list.c:585` | `llenCommand()` | `listTypeLength()` |
| LRANGE  | `src/t_list.c:875` | `lrangeCommand()` | `addListRangeReply()` |

---

## 1. LPUSH / RPUSH

**Source:** `src/t_list.c` — `lpushCommand()`, `rpushCommand()`, `pushGenericCommand()`

Both commands are thin wrappers over one generic function:

```c
void lpushCommand(client *c) { pushGenericCommand(c, LIST_HEAD, 0); }
void rpushCommand(client *c) { pushGenericCommand(c, LIST_TAIL, 0); }
```

`pushGenericCommand()` flow:
1. `lookupKeyWriteWithLink()` — fetch key or reserve a slot
2. `checkType()` — return WRONGTYPE error if not a list
3. Create a new listpack list if the key doesn't exist
4. `listTypeTryConversionAppend()` — check capacity **before** pushing, upgrade encoding if needed
5. Loop over all value arguments (`argv[2]` onwards) and push each
6. Return new length, fire keyspace event, increment dirty counter

### Best Practices

- **Single generic handler with a `where` flag** (`LIST_HEAD` / `LIST_TAIL`) eliminates duplication between LPUSH and RPUSH. The same pattern powers `LPUSHX`/`RPUSHX` via the `xx` parameter.
- **Pre-check capacity before inserting**, not after — avoids partial inserts in an inconsistent state.
- **Support multiple values in one call** — loop over all `argv[2..n]`, not just the first.
- **Always return the new length** after all pushes complete.

---

## 2. LLEN

**Source:** `src/t_list.c` — `llenCommand()` → `listTypeLength()`

```c
void llenCommand(client *c) {
    kvobj *kv = lookupKeyReadOrReply(c, c->argv[1], shared.czero);
    if (kv == NULL || checkType(c, kv, OBJ_LIST)) return;
    addReplyLongLong(c, listTypeLength(kv));
}
```

`listTypeLength()` dispatches by encoding:

```c
unsigned long listTypeLength(const robj *subject) {
    if (subject->encoding == OBJ_ENCODING_QUICKLIST)
        return quicklistCount(subject->ptr);
    else if (subject->encoding == OBJ_ENCODING_LISTPACK)
        return lpLength(subject->ptr);
    else
        serverPanic("Unknown list encoding");
}
```

### Best Practices

- **Encoding-agnostic wrapper** — the command handler never needs to know whether the list is a listpack or quicklist. All encoding details are hidden behind `listTypeLength()`.
- **Return 0 for non-existent key** — `lookupKeyReadOrReply` with `shared.czero` handles the missing-key case in one line.
- **Use pre-allocated shared responses** (`shared.czero`) to avoid allocations for common replies.

---

## 3. LRANGE

**Source:** `src/t_list.c` — `lrangeCommand()` → `addListRangeReply()`

```c
void lrangeCommand(client *c) {
    long start, end;
    if ((getLongFromObjectOrReply(c, c->argv[2], &start, NULL) != C_OK) ||
        (getLongFromObjectOrReply(c, c->argv[3], &end, NULL) != C_OK))
        return;

    kvobj *o = lookupKeyReadOrReply(c, c->argv[1], shared.emptyarray);
    if (!o || checkType(c, o, OBJ_LIST)) return;

    addListRangeReply(c, o, start, end, 0);
}
```

Range clamping in `addListRangeReply()`:

```c
// Step 1: convert negative indexes to absolute positions
if (start < 0) start = llen + start;
if (end < 0)   end   = llen + end;
if (start < 0) start = 0;           // still negative means before index 0

// Step 2: empty range check — bail early
if (start > end || start >= llen) {
    addReply(c, shared.emptyarray);
    return;
}

// Step 3: clamp upper bound
if (end >= llen) end = llen - 1;

rangelen = (end - start) + 1;
// dispatch to encoding-specific handler...
```

### Best Practices

- **Validate args before key lookup** — parse `start`/`end` first, return parse error before touching the DB.
- **Two-pass negative index resolution:**
  1. Convert negative → absolute (`llen + start`)
  2. Clamp if still negative (e.g. `-999` on a 3-element list still resolves to `0`)
- **Return empty array for out-of-range**, not an error — this is correct Redis semantics.
- **Clamp upper bound silently** (`end >= llen → llen - 1`) — don't error on over-large end indexes.
- **Delegate to encoding-specific handlers** only after all bounds are fully resolved.

---

## Dual Encoding: Listpack → Quicklist

**Source:** `src/t_list.c` — `listTypeTryConversionAppend()`, `listTypeTryConvertListpack()`, `listTypeTryConvertQuicklist()`

Redis stores lists in two formats depending on size:

| Encoding | Internal Structure | When Used |
|----------|--------------------|-----------|
| `OBJ_ENCODING_LISTPACK` | Compact sequential bytes | Default for new/small lists |
| `OBJ_ENCODING_QUICKLIST` | Linked list of listpack nodes | After exceeding `list-max-listpack-size` |

**Upgrade path (growing):** before each push, `listTypeTryConversionAppend()` checks if the incoming elements would push the listpack over the size/count limit. If so, it converts to a quicklist first.

**Downgrade path (shrinking):** after pops/deletes, `listTypeTryConvertQuicklist()` checks if the quicklist has shrunk to a single node small enough to become a listpack again. To avoid thrashing, the downgrade threshold is **half** the upgrade threshold.

### Best Practices

- **Start small, upgrade lazily** — new lists always begin as the compact format.
- **Convert before inserting**, not after — keeps the data structure valid at all times.
- **Use hysteresis on downgrade** (half the threshold) — prevents oscillating between encodings on repeated push/pop.
- **Abstract both encodings behind the same API** (`listTypePush`, `listTypeLength`, etc.) so callers are never aware of the internal format.

---

## General Patterns

| Pattern | Where in Redis | Benefit |
|---------|---------------|---------|
| Generic handler with direction flag | `pushGenericCommand(c, LIST_HEAD/TAIL, xx)` | One function for LPUSH, RPUSH, LPUSHX, RPUSHX |
| Encoding-agnostic wrappers | `listTypePush()`, `listTypeLength()` | Command code stays clean, encoding can change freely |
| Validate args → lookup key → type check → operate | All four commands | Fail fast, no wasted work |
| Pre-allocated shared replies | `shared.czero`, `shared.emptyarray` | Zero allocation for common responses |
| Negative index two-step normalization | `addListRangeReply()` | Correct behavior even for extreme negative indexes |
| Pre-conversion before insert | `listTypeTryConversionAppend()` | No invalid intermediate state |
