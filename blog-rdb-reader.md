# Building an RDB File Reader in Go: Parsing Redis Persistence From Scratch

*How we implemented a reader for Redis RDB dump files in Go—from the magic header to length-prefixed strings and key-value restoration.*

---

If you’ve ever wondered how Redis loads its snapshot on startup, you’ve run into **RDB** (Redis Database): Redis’s compact, binary dump format. In this post we walk through building an **RDB file reader in Go**—the same kind of logic a Redis-compatible server needs to restore state from a dump file.

We’ll cover the file layout, the opcode-based stream, and the slightly tricky **length-prefixed string encoding** that RDB uses. By the end, you’ll have a clear picture of how to parse RDB yourself.

---

## What is RDB?

RDB is Redis’s binary snapshot format. A single `.rdb` file contains:

- A fixed **header** (magic + version)
- Optional **auxiliary metadata**
- One or more **logical databases** (key-value pairs, with optional expiry)

The format is **opcode-driven**: special bytes (opcodes) tell the parser what kind of data comes next (e.g. “select DB”, “key with expiry in seconds”, “end of file”). Parsing is a loop: read opcode → interpret → read associated data → repeat until EOF or the end-of-file opcode.

---

## High-Level Design: The Reader

Our reader is a small struct that wraps a file and a buffered reader:

```go
type RDBReader struct {
    file   *os.File
    reader *bufio.Reader
}
```

We use `bufio.Reader` so we can read byte-by-byte and in fixed-size chunks without a syscall for every byte. The entry point opens the file, reads the header, then runs the main parsing loop.

---

## Step 1: The Header

The first 9 bytes of every RDB file are fixed:

- Bytes 0–4: magic string `"REDIS"`
- Bytes 5–8: version string (e.g. `"0011"`)

If either check fails, we reject the file. This is the reader’s “handshake” with the format.

```go
header := make([]byte, 9)
io.ReadFull(r.reader, header)
magicString := string(header[:5])   // "REDIS"
version := string(header[5:])      // "0011"
```

---

## Step 2: The Opcode Loop

After the header, the file is a stream of **opcodes** (single bytes) followed by their payloads. We define constants for the ones we care about:

| Opcode   | Constant   | Meaning                          |
|----------|------------|----------------------------------|
| `0xFF`   | `EOF`      | End of RDB file                  |
| `0xFE`   | `SELECTDB` | Next: database index and data    |
| `0xFD`   | `EXPIRETIME`  | Expiry in seconds (4 bytes)   |
| `0xFC`   | `EXPIRETIMEMS`| Expiry in milliseconds (8 bytes)|
| `0xFB`   | `RESIZEDB` | Hash table / expiry table sizes  |
| `0xFA`   | `AUX`      | Auxiliary key-value metadata     |

The main loop is:

1. Read one byte (opcode).
2. If EOF or opcode is `EOF` → stop.
3. Otherwise switch on opcode:
   - **AUX**: read two length-prefixed strings (key, value) and log or store metadata.
   - **SELECTDB**: call into “read database” logic (see below).

So the “shape” of the file is: **header**, then a sequence of **AUX** and **SELECTDB** sections until **EOF**.

---

## Step 3: Reading a Database (SELECTDB)

When we see `SELECTDB`, we:

1. Read the **database index** (one byte).
2. Then read a **sub-stream** that can contain:
   - **RESIZEDB** (`0xFB`): two bytes — hash table size and expiry table size (for sizing, we mainly log them).
   - **EXPIRETIME** (`0xFD`): 4-byte Unix timestamp in seconds (little-endian).
   - **EXPIRETIMEMS** (`0xFC`): 8-byte timestamp in milliseconds (little-endian).
   - **Value type 0** (string): key and value as length-prefixed strings; this is a plain key-value pair.

For each key-value pair we build an internal command (e.g. `SET key value` or `SET key value PX/EX expiry`) and feed it to the command processor so the in-memory store is restored. Expiry is attached when we’ve just read an expiry opcode before the key-value.

---

## Step 4: Length-Prefixed Strings — The Tricky Part

RDB doesn’t use null-terminated strings. It uses **length-prefixed encoding**: the first byte (or more) tells you how many bytes the string is. The twist is that the **first byte is split into a 2-bit type and 6 bits of length**:

- **First 2 bits** (prefix type):
  - `00`: length is in the **last 6 bits** (0–63 bytes).
  - `01`: length is **14 bits** (next 6 bits + next byte, big-endian).
  - `10`: length is **32 bits** (next 4 bytes, big-endian).
  - `11`: **encoded integer string** — the next 6 bits choose the format (e.g. 8-, 16-, or 32-bit integer), and we read that many bytes and turn them into a string representation (e.g. for Redis’s integer-encoded values).

So one byte can mean “short string”, “medium string”, “long string”, or “integer as string”. In code it looks like:

```go
prefixBits := (prefix & 0xC0) >> 6   // top 2 bits
remainingBits := prefix & 0x3F       // low 6 bits
switch prefixBits {
case 0:  // 6-bit length
    length := int(remainingBits)
    buffer, _ := r.ReadNBytes(length)
    result = string(buffer)
case 1:  // 14-bit length (remaining 6 bits + next byte)
    // ... read 2 bytes, big-endian uint16
case 2:  // 32-bit length
    // ... read 4 bytes, big-endian uint32
case 3:  // integer encoding (8/16/32 bit)
    // ... read 1, 2, or 4 bytes, little-endian, format as string
}
```

Getting this right is the core of parsing keys, values, and metadata in RDB.

---

## Takeaways

- **RDB is binary and opcode-based**: after a 9-byte header, you have a stream of opcodes and their payloads until EOF or the EOF opcode.
- **Opcodes** tell you what to read next: auxiliary metadata, database selection, table sizes, expiry, or key-value pairs.
- **Length-prefixed strings** use the first byte’s top 2 bits to decide between 6-bit, 14-bit, 32-bit length or integer encoding—efficient and compact.
- **Endianness matters**: RDB uses **little-endian** for numeric payloads (timestamps, integer strings) and **big-endian** for multi-byte lengths in the 14- and 32-bit cases.

Implementing this in Go with `bufio.Reader`, `io.ReadFull`, and `binary.LittleEndian`/`BigEndian` gives you a clean, testable RDB reader that can drive key-value restoration (and optionally expiry) in a Redis-compatible server.

---

*This reader is part of a Redis-compatible cache server built for learning Redis internals. You can use the same ideas to build tooling, debug dumps, or your own persistence layer.*
