# MemStore

- in-memory heart of kivi.
- every read-write will go through it.
- it's essentially a thread-safe map 
    - but designed cleanly to make other stores build upon it without rewriting
- it's thread-safe because multiple goroutines can call `Get`, `Set`, `Delete` simultaneously.
- it' build using Go Generics, so that it can be used with different types.

## Mental Model
```text
MemStore
    sync.RWMutex
        map[string]entry[V]
```
- each key maps to an `entry`.
    - `entry` holds the Value and optional expiry time.

## RWMutex
- it has 2 modes:
    - read lock: multiple go-routines can hold it simulatenously 
    - write lock: only 1 go-routine can hold it and it blocks other readers and writers
- Rule:
    - `Get`: ReadLock -> Read -> ReadUnlock
    - `Set`/`Delete`: Lock -> Update -> Unlock
