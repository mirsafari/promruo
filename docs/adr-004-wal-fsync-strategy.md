# ADR-004: WAL fsync Strategy — Sync on Close Only

**Status:** Accepted (with known tradeoff)  
**Phase:** 1 (Storage Engine)

## Context

The WAL is an append-only file. Every `Append` call writes 48 bytes to disk. The question is: when do we call `fsync` (or `fdatasync`) to guarantee those bytes are durable — i.e., survive a process crash or power loss?

The OS write path looks like this:

```
Append() → write() syscall → OS page cache → [fsync] → physical disk
```

Without `fsync`, the OS can hold written data in its page cache for seconds or minutes before flushing to disk. A crash during that window loses the data even though `write()` returned successfully.

## Decision

Call `Sync()` only on `Close()`. Individual `Append` calls do not fsync.

## Why not fsync per write?

A single `fdatasync` on a modern SSD takes approximately **1–5ms** (it flushes the drive's write buffer). This caps throughput at **200–1000 writes/second** — a hard ceiling imposed by the storage hardware, not by Go or the CPU.

For a system that ingests one PromQL rollup result per scheduled interval (weekly, monthly), this limit is irrelevant. But the tradeoff is real and worth understanding.

## The production solution: group commit

High-throughput WALs (PostgreSQL, Kafka, etcd/bbolt) use **group commit**:

1. Multiple goroutines call `Append` concurrently.
2. Writes are buffered.
3. One fsync is issued for the entire batch.
4. All callers in the batch are acknowledged together.

This amortizes the fsync cost across N writes instead of paying it N times. Kafka achieves hundreds of thousands of writes/second this way.

## Crash recovery implications

With sync-on-close only:
- Entries written and `Close()`-d are durable.
- Entries written but not yet `Close()`-d may be lost on crash.
- On restart, `OpenWAL` reads the existing file size — any partial/lost entries result in a shorter-than-expected WAL, which is detectable but not automatically corrected.

A production WAL would add a **CRC checksum** per record so that partial writes (e.g., 24 of 48 bytes written before crash) are detected and truncated on recovery rather than silently producing corrupt entries.

## Consequences

- Throughput is limited only by write syscall overhead (~millions/sec), not fsync latency.
- A crash between `Append` and `Close` loses the un-fsynced entries.
- This is an acceptable tradeoff for a low-frequency scheduled rollup system.
- If durability guarantees become a requirement, the fix is: call `w.f.Sync()` after each `Append`, or implement group commit.
