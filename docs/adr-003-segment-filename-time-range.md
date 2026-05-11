# ADR-003: Segment Filenames Encode Data Time Range

**Status:** Accepted  
**Phase:** 1 (Storage Engine)

## Context

When `ScanSegments` receives a time-range query, it needs to know which segment files contain relevant data. The original implementation loaded every entry from every `.seg` file into memory and then filtered — O(n) reads across all data regardless of query range.

The question was: how can we skip entire segment files that can't possibly contain matching data?

## Decision

Encode the **min and max timestamp of the records inside** directly into the segment filename:

```
{minTimestamp}-{maxTimestamp}.seg
```

Example: `1715155200000000000-1715758800000000000.seg`

Both timestamps are Unix nanoseconds, left-padded to fixed width by Go's default int64 formatting (no padding needed — all nanosecond timestamps since 2001 are 19 digits).

The flush path (`Flusher`) derives min/max from the sorted entries before writing:

```go
minTs := sorted[0].Timestamp
maxTs := sorted[len(sorted)-1].Timestamp
name := fmt.Sprintf("%d-%d.seg", minTs, maxTs)
```

`ScanSegments` parses the filename and skips any file where the segment's range does not overlap the query range:

```go
// overlap condition: seg.min <= query.end && seg.max >= query.start
```

## Why not use creation timestamp?

The previous filename was `time.Now().UnixNano().seg` — the creation time of the segment, not the time range of the data. This gives no useful information for filtering: a segment created at time T could contain data from any arbitrary range.

## Rationale

- **Block-level filtering:** Entire files are skipped without opening them. For a query covering 1 week of data in a system with years of segments, the majority of files are eliminated at the filename-parse stage.
- **No sidecar files:** Encoding range in the filename avoids the complexity of maintaining a separate metadata/index file that must be kept in sync.
- **Self-describing files:** A human or tool can determine a segment's data range from the filename alone — useful for debugging and manual inspection.
- **This is how real engines work:** Prometheus TSDB, LevelDB, and RocksDB all store block boundary metadata to enable skipping. This is a simplified version of the same idea.

## Consequences

- Segment filenames are longer and less human-readable than a single timestamp.
- Flushing an empty WAL is now explicitly rejected (no min/max to derive).
- Parsing the filename adds a small amount of code (`segmentTimeRange` helper).
- Existing `.seg` files from the old naming scheme are incompatible — they will be skipped by `isSegmentFile` since they contain only one timestamp component (a breaking change, acceptable at this stage).
