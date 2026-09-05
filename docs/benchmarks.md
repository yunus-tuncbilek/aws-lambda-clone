# Benchmarks

Rough numbers captured as we build (methodology gets rigorous in Phase 6:
scripted load generator, percentiles, stable host). Environment: Colima VM
(Apple Virtualization.Framework), containerd 2.3.4, arm64, `hello` = 8 MB
scratch image, measured from inside the VM.

## Cold start (single sequential invocations)

| Phase | Isolation | Latency | Notes |
|---|---|---|---|
| 1 | subprocess (no isolation) | ~15 ms | process spawn + readiness poll |
| 2 | container per invocation | ~65 ms (51–82) | + containerd RPC, overlay snapshot, ns/cgroup setup, userland boot |

**Container overhead vs. subprocess: ~50 ms.** This is the gap the Phase 4 warm
pool targets (reuse a running container → skip all of the above).

## Still to measure
- Phase 3: concurrent-request ceiling before latency/error blowup.
- Phase 4: warm-invocation latency vs. cold start (headline comparison).
- Phase 5 (stretch): container vs. Firecracker microVM.
