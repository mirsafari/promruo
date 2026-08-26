# ADR-005: Operator Component Architecture

**Status:** Accepted  
**Phase:** 2 (Operator Logic)

## Context

The operator needs to watch `MetricRollup` custom resources and act on them by fetching Prometheus data and persisting rollup results to the storage engine. Several architectural shapes are possible: a single monolithic binary, a controller with internal goroutines for scheduling, or a controller that delegates work to Kubernetes primitives.

The key constraint is that the storage engine writes to a Persistent Volume, which in most clusters supports only `ReadWriteOnce` — a single node can mount it at a time. Any design where multiple pods write directly to the PVC risks mount conflicts or WAL corruption.

## Decision

Split the system into three logical components, each compiled as a separate entrypoint from the same container image:

### 1. Controller (`cmd/controller`) — Kubernetes Deployment

Stateless. No PVC. Implements the `controller-runtime` reconcile loop for `MetricRollup` CRDs.

On each reconcile it ensures:

- A **Storage StatefulSet** exists (creates or updates it).
- A **Kubernetes CronJob** exists for each `MetricRollup`, with the schedule derived from `spec.interval` and the Storage API URL injected as an environment variable.
- Owner references are set so that deleting a `MetricRollup` cascades to its CronJob. The Storage StatefulSet is cluster-scoped and shared, so it is not owned by any single `MetricRollup`.

The Controller does not track `LastExecutionTime` or `NextExecutionTime` in the `MetricRollupStatus` — Kubernetes CronJob status is the authoritative source for scheduling history.

### 2. Storage Server (`cmd/storage`) — Kubernetes StatefulSet, 1 replica

Owns the Persistent Volume. Runs the WAL and segment engine from Phase 1 behind an API.

Minimum endpoints:

- `write` — accepts a `Entry`, appends to WAL.
- `query` — accepts `metricHash`, `start`, `end` query parameters, returns matching entries (Phase 4).

Single replica enforces the single-writer constraint. No horizontal scaling at this stage.

### 3. CronJob Worker (`cmd/cronjob`) — Kubernetes CronJob, ephemeral

Runs to completion on each scheduled tick. Steps:

1. Read `MetricRollup` spec fields from environment variables (injected by Controller when creating the CronJob).
2. Execute PromQL range query against Prometheus.
3. Downsample the result (mean or median, depending on spec).
4. Write the result to the Storage HTTP API.
5. Exit 0 on success, non-zero on failure (Kubernetes will mark the job failed).

Retries on Storage API failures use **exponential backoff**, with the total maximum retry duration set below the CronJob's `activeDeadlineSeconds`. This ensures the pod terminates cleanly rather than being killed mid-write.

## Rationale

**Controller as Deployment (stateless):** The reconcile loop only reads CRD state and creates/updates other Kubernetes resources. It has no affinity to a particular node or disk. A Deployment is the correct workload type for stateless control-plane logic.

**Storage as StatefulSet:** The storage engine writes to a local PVC. A StatefulSet provides a stable pod identity and a stable PVC claim. When the pod is rescheduled, it reconnects to the same volume. A Deployment could work but does not provide stable volume binding guarantees.

**Kubernetes CronJob instead of internal goroutines:** Delegating scheduling to Kubernetes CronJobs means the Controller is stateless between reconcile calls. The alternative is an internal `time.Ticker` or `robfig/cron` scheduler inside the Controller - requires the Controller to survive pod restarts without missing ticks. Kubernetes CronJobs have built-in `startingDeadlineSeconds` and `concurrencyPolicy` to handle missed schedules and overlapping runs natively.

**API between CronJob and Storage (not direct PVC mount):** CronJob pods are ephemeral and may land on any node. Most clusters use `ReadWriteOnce` PVCs. Mounting the same PVC from two nodes simultaneously is either impossible or results in undefined behavior.

**Same image, different entrypoints:** Building three separate images adds CI complexity with no benefit at this scale. A single multi-entrypoint image is built once, tested once, and the entrypoint is selected via the container command in each workload spec.

## Consequences

- The Controller must manage the lifecycle of two downstream resource types (Storage StatefulSet + CronJobs) per `MetricRollup`. Reconcile logic is more involved than a single-resource controller.
- The Storage HTTP API must be implemented before CronJob workers can be tested end-to-end (Phase 3 depends on Phase 2 completing the write endpoint).
- `MetricRollupStatus` carries less information than originally described — scheduling state lives in the CronJob, not the CR status. The status subresource can still surface a human-readable summary (last successful run, storage health).
- Horizontal scaling of storage requires a different design (distributed WAL, leader election, or an external store). Out of scope for this project.
