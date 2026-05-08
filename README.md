# promruo

Prometheus Rollup Operator is a lightweight, self-sufficient Kubernetes Operator designed to distill high-resolution Prometheus metrics into long-term, low-resolution data points. It is built for weekly/monthly/yearly trend analysis without the operational overhead of a full-scale long-term storage solution like Thanos, Cortex or Mimir.

> [!CAUTION] 
> **Promruo is an experimental project.** This tool is intended for learning purposes on how TSDB and Kubernetes operator work under the hood. It does not provide the data redundancy, horizontal scaling, or rigorous consistency guarantees of production-grade storage backends.

**What it is:**

- A Precision Tool: Purpose-built to execute specific PromQL queries on a schedule (weekly/monthly) and store the aggregated results (mean, median, etc.).
- Self-Sufficient: A single Go binary that includes a custom, append-only binary storage engine, so no external databases (Postgres, Redis) required.
- Kubernetes Native: Managed via Custom Resource Definitions (CRDs); simply define a MetricRollup object and let the operator handle the rest.
- Disk Efficient: Uses a fixed-size binary format and Write-Ahead Logging (WAL) to ensure years of data take up only megabytes of space.
- Local-First: Designed to run as a StatefulSet with a Persistent Volume, keeping your archival data right where your cluster is.

**What it is NOT:**

- A TSDB Replacement: This is not a replacement for Prometheus, Thanos, or VictoriaMetrics. It does not support high-cardinality raw data or sub-minute scraping.
- A Visualization Suite: While it includes a minimal embedded UI for trend checking, it is not a replacement for Grafana.
- A Real-Time Alerting Engine: Promruo is for historical analysis, not for triggering real-time incident responses.
- High Availability (by default): Since it relies on local disk persistence and a single-binary storage engine, it prioritizes simplicity over distributed consensus.
