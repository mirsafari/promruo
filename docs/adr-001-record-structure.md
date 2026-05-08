# ADR-001: Fixed-Size Binary Record Format

**Status:** Accepted  
**Phase:** 1 (Storage Engine)

## Context

The storage engine needs to persist PromQL rollup results (timestamp + value pairs) to disk. We want a format that is simple to implement, enables efficient reads, and requires no external dependencies.

## Decision

Use a fixed-length 48-byte binary record with the following layout:

| Offset | Size | Field        | Type      |
|--------|------|-------------|-----------|
| 0      | 8    | Timestamp   | int64     |
| 8      | 8    | Value       | float64   |
| 16     | 32   | MetricHash  | [32]byte  |

All multi-byte values are encoded in **little-endian** order.

## Rationale

- **O(1) disk seeks:** Because every record is exactly 48 bytes, reading record N requires seeking to offset `N * 48`. Variable-length formats (JSON, protobuf, CSV) require scanning or a separate index.
- **Timestamp as int64:** Unix timestamp in nanoseconds covers a range well beyond any practical use case (until year 2262). Sorting by this field enables efficient range scans on sorted segments.
- **Value as float64:** IEEE 754 double precision is sufficient for Prometheus metric values. A fixed-size float avoids varint encoding overhead.
- **MetricHash as [32]byte:** SHA-256 hash of the metric name. 32 bytes provides collision resistance without storing variable-length strings. Filtering during reads is done by hash comparison rather than string comparison.
- **No framing:** Since records are fixed-size, there is no need for length prefixes, delimiters, or framing markers. The file can be read as a flat array of records.
- **No external dependencies:** The standard library `encoding/binary` package is sufficient. No protobuf, FlatBuffers, or other serialization framework is needed.

## Consequences

- Record size is fixed at 48 bytes — approximately 20,000 records per MB of disk.
- Changing the record layout in the future would require a migration or versioning scheme.
- Metric names longer than 32 bytes must be truncated before hashing (though SHA-256 output is always 32 bytes; the input metric name can be any length).
