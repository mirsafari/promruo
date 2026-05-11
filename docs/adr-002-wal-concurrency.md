# ADR-002: WAL Concurrency Model — Mutex vs Atomic

**Status:** Accepted  
**Phase:** 1 (Storage Engine)

## Context

The WAL (`internal/storage/wal.go`) tracks its current write offset in `currentSize int64`. Multiple goroutines may call `Append()` and `Size()` concurrently. The question was: how do we protect `currentSize` from data races?

Two candidates:
- `sync.Mutex` — a lock that serializes all access
- `sync/atomic` (`atomic.Int64`) — CPU-level memory barriers that make a single read or write indivisible

## Decision

Use `atomic.Int64` for `currentSize`. Use `sync.Mutex` only for compound operations (`Append`, `Flusher`).

```go
type WAL struct {
    f           *os.File
    path        string
    currentSize atomic.Int64  // lock-free single-field access
    mu          sync.Mutex    // guards compound operations
}
```

`Size()` reads without acquiring the mutex:

```go
func (w *WAL) Size() int64 {
    return w.currentSize.Load()
}
```

`Append` still holds the mutex because it is a compound operation (read offset → write at offset → increment offset):

```go
w.mu.Lock()
defer w.mu.Unlock()
bytesWritten, err := w.f.WriteAt(data, w.currentSize.Load())
w.currentSize.Add(int64(bytesWritten))
```

## Why atomics at all?

A plain `int64` read/write is a single CPU instruction on 64-bit platforms and *usually* works without synchronization. But two problems remain:

1. **32-bit platforms:** An 8-byte read/write takes two instructions. Another goroutine can observe a half-written value between them (a "torn read").
2. **Go memory model:** Even on 64-bit, the compiler and CPU are permitted to reorder instructions and cache values in registers across goroutine boundaries. A plain variable write in goroutine A is not guaranteed to be visible in goroutine B without an explicit synchronization point. `atomic.Int64` emits a **memory barrier** — a CPU instruction that flushes in-flight writes and prevents reordering around the atomic operation.

Without `atomic` or a mutex, the race detector will flag `Size()` as a data race.

## Why not just use the mutex in `Size()` too?

It would be correct. The reason to prefer atomics here is that `Size()` is a single-field read with no compound invariant — it doesn't need to be consistent with any other field. A mutex would add unnecessary contention for callers that just want to check the WAL size (e.g., to decide whether to flush). Atomics are the right tool when the operation is truly indivisible by nature.

## The key distinction: atomic vs mutex

Use **atomic** when:
- The operation is a single read or write on one value
- No other state must be consistent with it at the same time

Use **mutex** when:
- Multiple fields must be consistent with each other
- The operation is compound: read-then-decide-then-write

`Flusher` is a clear mutex case: it reads all WAL entries, writes a segment file, then resets `currentSize` to 0. All three steps must appear as one uninterrupted unit. If `Size()` returned 0 halfway through a flush, a caller might incorrectly decide not to flush again.

## Consequences

- `Size()` is lock-free — safe to call from any goroutine without contention
- The race detector accepts the implementation without warnings
- `atomic.Int64` (value type, not pointer) cannot be copied after first use — the WAL must always be passed as a pointer, which it already is
- The distinction between "atomic for single reads" and "mutex for compound operations" is a reusable mental model for all future concurrent data structures in this project
