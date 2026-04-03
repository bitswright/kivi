# Kivi 🪨

> A persistent key-value store built from scratch in Go — designed as a learning project to understand how real-world storage engines work under the hood.

Kivi is not meant to replace Redis or RocksDB. It's meant to help *understand* them. By building each layer — from an in-memory map to a WAL, log compaction, LSM Trees, and transactions — genuine intuition for the design decisions behind every serious database is developed.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [API Reference](#api-reference)
- [Build & Run](#build--run)
- [Task List](#task-list)
- [Running Tests](#running-tests)
- [Concepts & Real-World Parallels](#concepts--real-world-parallels)
- [References](#references)

---

## Features

- **In-memory store** backed by a `map[string]string` with `sync.RWMutex` for concurrent access
- **Append-only log** for persistence — replayed on startup to reconstruct state
- **Write-Ahead Log (WAL)** with `fsync` and CRC32 checksums for crash safety
- **Segmented log + Bitcask model** — in-memory hash index over fixed-size log segments with background compaction
- **LSM Tree engine** — MemTable flushed to SSTables on disk, with Bloom filters and multi-level compaction
- **TTL / key expiry** — lazy and active expiry strategies
- **Atomic batch writes** — all-or-nothing semantics over multiple operations
- **Range scan API** — lexicographic key range queries powered by sorted SSTables
- **HTTP API** — simple REST interface over all store operations

---

## Architecture

Kivi is built in progressive layers, each adding a new capability on top of the previous one:

```
┌─────────────────────────────────┐
│         HTTP API Layer          │  ← REST interface (GET, PUT, DELETE)
├─────────────────────────────────┤
│         Transaction Layer       │  ← Atomic batch writes, MVCC, TTL
├─────────────────────────────────┤
│         LSM Tree Engine         │  ← MemTable + SSTables + Bloom filters
├─────────────────────────────────┤
│     Compaction + Segmented Log  │  ← Bitcask model, hint files
├─────────────────────────────────┤
│      WAL + Crash Recovery       │  ← fsync, CRC32 checksums
├─────────────────────────────────┤
│       Append-Only Log           │  ← Durability, log replay on startup
├─────────────────────────────────┤
│     In-Memory Store (map)       │  ← RWMutex, concurrent access
└─────────────────────────────────┘
```

All storage backends implement a common `Store` interface, so the HTTP layer is completely decoupled from the underlying engine:

```go
type Store interface {
    Get(key string) (string, error)
    Set(key, value string) error
    Delete(key string) error
    Keys() []string
}
```

---

## Project Structure

```
kivi/
├── cmd/
│   └── kivi/
│       └── main.go           # Entry point — starts the HTTP server
├── internal/
│   ├── store/
│   │   ├── store.go          # Core Store interface
│   │   ├── memstore.go       # In-memory map-backed store
│   │   ├── logstore.go       # Append-only log store
│   │   ├── walstore.go       # WAL with fsync + CRC32
│   │   ├── segstore.go       # Segmented log + Bitcask index
│   │   └── lsmstore.go       # LSM Tree (MemTable + SSTables)
│   ├── wal/
│   │   ├── wal.go            # Write-Ahead Log implementation
│   │   └── entry.go          # Log entry encoding/decoding + checksum
│   ├── lsm/
│   │   ├── memtable.go       # In-memory sorted structure
│   │   ├── sstable.go        # Sorted String Table (on-disk)
│   │   ├── compactor.go      # Background compaction logic
│   │   └── bloom.go          # Bloom filter per SSTable
│   ├── segment/
│   │   ├── segment.go        # Log segment management
│   │   └── index.go          # In-memory hash index (Bitcask model)
│   └── http/
│       ├── server.go         # HTTP server setup
│       └── handlers.go       # Route handlers for all API endpoints
├── tests/
│   ├── crash_test.go         # Crash recovery tests
│   ├── concurrent_test.go    # Race condition tests
│   ├── compaction_test.go    # Compaction correctness tests
│   └── ttl_test.go           # TTL expiry tests
├── data/                     # Runtime data directory (gitignored)
│   ├── segments/             # Log segment files
│   └── sstables/             # SSTable files
├── go.mod
├── go.sum
└── README.md
```

---

## API Reference

All endpoints communicate over HTTP. Default port is `:5000`.

### `PUT /:key`
Set a value for a key.

**Request body:**
```json
{
  "value": "hello world",
  "ttl": 3600
}
```
- `value` (string, required) — the value to store.
- `ttl` (int, optional) — time-to-live in seconds. Omit for no expiry.

**Response:** `200 OK`
```json
{ "ok": true }
```

---

### `GET /:key`
Retrieve a value by key.

**Response:** `200 OK`
```json
{ "key": "username", "value": "harish" }
```

**Not found:** `404 Not Found`
```json
{ "error": "key not found" }
```

---

### `DELETE /:key`
Delete a key.

**Response:** `200 OK`
```json
{ "ok": true }
```

---

### `GET /keys`
List all active keys.

**Response:** `200 OK`
```json
{ "keys": ["username", "session_id", "theme"] }
```

---

### `GET /range?start=a&end=z`
Returns all key-value pairs within a lexicographic key range. Powered by the sorted SSTable layout — available from the LSM Tree milestone onwards.

**Response:** `200 OK`
```json
{
  "pairs": [
    { "key": "apple", "value": "fruit" },
    { "key": "avocado", "value": "also fruit" }
  ]
}
```

---

### `POST /batch`
Atomically apply multiple write operations. All succeed or none do.

**Request body:**
```json
{
  "ops": [
    { "op": "set", "key": "a", "value": "1" },
    { "op": "set", "key": "b", "value": "2" },
    { "op": "del", "key": "c" }
  ]
}
```

**Response:** `200 OK`
```json
{ "ok": true, "applied": 3 }
```

---

## Build & Run

**Prerequisites:** Go 1.21+

```bash
git clone https://github.com/your-username/kivi.git
cd kivi
go run ./cmd/kivi
```

The server starts on `http://localhost:5000`.

**Quick test with curl:**
```bash
# Set a key
curl -X PUT http://localhost:5000/username \
  -H "Content-Type: application/json" \
  -d '{"value": "harish"}'

# Get a key
curl http://localhost:5000/username

# Delete a key
curl -X DELETE http://localhost:5000/username

# List all keys
curl http://localhost:5000/keys

# Range scan
curl "http://localhost:5000/range?start=a&end=z"

# Batch write
curl -X POST http://localhost:5000/batch \
  -H "Content-Type: application/json" \
  -d '{"ops": [{"op":"set","key":"x","value":"1"},{"op":"set","key":"y","value":"2"}]}'
```

---

## Task List

### Phase 1 — In-Memory Store + HTTP API
- [ ] Define the `Store` interface in `internal/store/store.go`
- [ ] Implement `MemStore` backed by `map[string]string` with `sync.RWMutex`
- [ ] Set up the HTTP server in `internal/http/server.go`
- [ ] Implement `PUT /:key`, `GET /:key`, `DELETE /:key`, `GET /keys` handlers
- [ ] Write concurrent access tests with `go test -race`
- [ ] Manual end-to-end test with `curl`

### Phase 2 — Append-Only Log
- [ ] Implement `LogStore` — wraps `MemStore` and appends every write to a log file
- [ ] Define log entry format: `OP key [value]\n`
- [ ] Open the log file in append mode (`O_APPEND | O_CREATE`)
- [ ] Implement `replayLog()` — reconstruct state from log on startup
- [ ] Handle malformed or incomplete entries gracefully during replay
- [ ] Write a restart test: populate keys → restart → verify keys are present
- [ ] Benchmark write throughput with and without the log

### Phase 3 — WAL + Crash Safety
- [ ] Implement `WAL` in `internal/wal/wal.go` with binary entry encoding
- [ ] Add CRC32 checksum to each log entry for corruption detection
- [ ] Call `file.Sync()` (fsync) after every write
- [ ] Implement WAL replay that detects and skips corrupted tail entries
- [ ] Add configurable `SyncMode`: `SyncAlways` (durable) vs `SyncNever` (fast, lossy)
- [ ] Write a crash simulation test: `kill -9` mid-write → restart → verify clean recovery
- [ ] Benchmark the performance cost of fsync vs no fsync

### Phase 4 — Compaction + Segmented Logs
- [ ] Implement `Segment` with a configurable max size (e.g. 1MB)
- [ ] Implement segment rotation: new segment created when active segment is full
- [ ] Implement in-memory `HashIndex`: `map[string]IndexEntry` storing `(file, offset, size)`
- [ ] Update `Set` to write to the active segment and update the index
- [ ] Update `Get` to look up the index and seek directly to the value on disk
- [ ] Implement background `Compactor`: merge old segments → new segment → update index
- [ ] Write hint files alongside compacted segments for fast index reconstruction on startup
- [ ] Write a compaction test: 100k writes → compaction → verify disk usage is bounded
- [ ] Verify reads are not blocked during background compaction

### Phase 5 — LSM Tree
- [ ] Implement `MemTable` using a sorted structure (`btree` or skip list)
- [ ] Implement MemTable size tracking and flush trigger
- [ ] Implement `SSTable` writer: sorted key-value pairs with a trailing index block
- [ ] Implement `SSTable` reader: binary search on the index block, seek to value
- [ ] Add a WAL to protect the MemTable from crash loss
- [ ] Implement per-SSTable Bloom filter in `internal/lsm/bloom.go`
- [ ] Implement multi-level compaction in `internal/lsm/compactor.go`
- [ ] Implement `LSMStore` composing all layers behind the `Store` interface
- [ ] Add `GET /range` endpoint leveraging sorted SSTable layout
- [ ] Write range scan test: 10k inserts → verify range query returns correct sorted results
- [ ] Benchmark read/write throughput vs the Bitcask (Phase 4) implementation

### Phase 6 — Transactions + Stretch Goals
- [ ] Implement `POST /batch` with all-or-nothing semantics (hold write lock for full batch)
- [ ] Implement rollback: if any op in the batch fails, undo all previous ops
- [ ] Add `ttl` field to log entries and MemTable entries
- [ ] Implement lazy expiry: check TTL on every `Get`, return 404 and delete if expired
- [ ] Implement active expiry: background goroutine with configurable sweep interval
- [ ] Write TTL test: set key with 1s TTL → sleep 2s → verify key is gone
- [ ] Write batch atomicity test: valid ops + one invalid op → verify zero partial writes
- [ ] *(Stretch)* Add a follower node with leader-to-follower log shipping over HTTP
- [ ] *(Stretch)* Replace HTTP with a custom binary TCP protocol (inspired by Redis RESP)
- [ ] *(Stretch)* Replace HTTP handlers with a gRPC service using a `.proto` definition
- [ ] *(Stretch)* Profile the server under load with `pprof` and fix the top bottleneck

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector — always do this
go test -race ./...

# Run benchmarks
go test -bench=. ./...

# Run a specific test file
go test -run TestCrashRecovery ./tests/

# Simulate a crash (server must be running)
go run ./cmd/kivi &
kill -9 $(pgrep kivi)
go run ./cmd/kivi   # restart and verify recovered state
```

---

## Concepts & Real-World Parallels

| Kivi | Production system |
|---|---|
| Append-only log | Kafka, PostgreSQL WAL |
| WAL + fsync | PostgreSQL, MySQL InnoDB |
| Bitcask model | Riak |
| LSM Tree | RocksDB, LevelDB, Cassandra, TiKV |
| MemTable + SSTable | LevelDB, Cassandra |
| Bloom filters | Cassandra, HBase, RocksDB |
| MVCC | PostgreSQL, CockroachDB, TiKV |
| TTL + active expiry | Redis volatile keys |
| Log shipping | PostgreSQL streaming replication |

---

## References

- [Designing Data-Intensive Applications — Martin Kleppmann](https://dataintensive.net/) — Chapter 3 covers LSM Trees and B-Trees in depth.
- [Bitcask: A Log-Structured Hash Table for Fast Key/Value Data](https://riak.com/assets/bitcask-intro.pdf) — The original Bitcask paper. Short and very readable.
- [The Log — Jay Kreps](https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying) — Why the append-only log is the common abstraction behind databases, queues, and distributed systems.
- [LevelDB Implementation Notes](https://github.com/google/leveldb/blob/main/doc/impl.md) — How the LSM Tree is implemented in LevelDB.
- [Redis Persistence Docs](https://redis.io/docs/manual/persistence/) — RDB vs AOF, directly comparable to the persistence models built here.