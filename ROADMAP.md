# Roadmap: FaaS Platform (AWS Lambda Clone)

Goal per [CLAUDE.md](CLAUDE.md): go **deep** on 4–5 hard distributed-systems
problems, ship a smaller well-documented version with **real benchmark numbers**,
deploy it, and write it up. Every phase below ends in a measurable, demoable
state so you're never far from something showable.

---

## Key decisions to lock first

### Language — recommend **Go**
The entire reference stack (faasd, OpenFaaS, firecracker-containerd, containerd,
`firecracker-go-sdk`) is Go. First-class Docker/containerd/Firecracker SDKs,
trivial concurrency for the scheduler, single-binary deploys. This is the
lowest-friction path to the microVM stretch goal and the most credible language
choice to defend in an interview for *this* domain. (Rust is viable but you'd
fight thinner SDKs; Python/Node would force you to shell out for everything and
undercut the systems narrative.)

### Dev environment — you need Linux
This mac (arm64, no Docker) can't do the real work directly:
- **Containers:** run a Linux VM locally — [Colima](https://github.com/abiosoft/colima)
  or [Lima](https://github.com/lima-vm/lima) gives you a Docker/containerd
  daemon on the mac. Fine for Phases 1–4.
- **Firecracker (Phase 5):** needs **KVM**, which nested virt on a mac does not
  reliably provide. Plan on a cloud **bare-metal / metal instance** (e.g. AWS
  `*.metal`, or a GCP instance with nested virt, or Equinix Metal). Budget for
  this only when you reach Phase 5.
- Benchmarks should run on a **stable Linux host**, not the mac, so numbers are
  reproducible and defensible.

---

## Phase 0 — Scaffolding & contracts (foundation)
Define the shapes everything else depends on before writing execution logic.
- Repo layout, Go module, Makefile, lint/test CI.
- **Function contract:** what *is* a function? Decide the handler interface
  (recommend: container reads a JSON event on stdin / an HTTP request on a local
  port, returns a JSON response). Keep it language-agnostic.
- **Control-plane API:** `POST /functions` (register), `POST /invoke/{name}`
  (run), `GET /functions/{name}` (status). HTTP/JSON.
- **Benchmark harness stub** (do this early, grow it every phase): a load
  generator (`vegeta`/`k6`/`wrk`) + a script that records p50/p95/p99 latency
  and error rate to a file. Numbers are the deliverable — make them cheap to
  collect from day one.
- **Exit criteria:** `POST /invoke` returns a hardcoded response; harness prints
  latency percentiles.

## Phase 1 — Synchronous execution MVP (no isolation)
Prove the event → run → response loop end to end.
- Executor runs the function as a **subprocess** (simplest possible backend).
- Wire the function contract through: event in, response out, errors/timeouts
  handled.
- **Benchmark:** baseline invocation latency (this is your theoretical floor).
- **Exit criteria:** deploy a real "hello" function and invoke it over HTTP.

## Phase 2 — Container-based isolation per invocation *(scope item 3)*
Swap the subprocess for a fresh container per request.
- Use the **Docker API** (`docker/docker/client`) first for speed, or go
  straight to **containerd** if you want the lower-level story. Recommend
  starting with Docker API, keeping the executor behind an interface so you can
  swap backends.
- Build **base runtime images** (one per supported language; start with one).
- Handle image pull/caching, container lifecycle, log/stdout capture, timeouts,
  cleanup on failure.
- **Benchmark → headline number #1:** cold-start latency (container spin-up +
  first invocation). Break it down (create / start / exec) so you can explain it.
- **Exit criteria:** each invocation runs in its own container; cold-start
  numbers recorded.

## Phase 3 — Scheduler / queue & concurrency *(scope item 2)*
Turn a single executor into a system that handles concurrent load without
falling over.
- **Request queue** + bounded **worker pool**; per-function and global
  concurrency limits.
- **Admission control / backpressure:** shed or queue when saturated; surface
  queue depth. Decide the failure mode (429 vs. queue-and-wait) deliberately —
  this is an interview talking point.
- Metrics: queue depth, in-flight count, wait time vs. exec time.
- **Benchmark → headline number #3:** concurrent-request ceiling before latency
  blows up / errors start. Plot latency vs. concurrency.
- **Exit criteria:** system degrades gracefully under load; ceiling documented.

## Phase 4 — Warm-pool reuse *(scope item 4 — the core of the project)*
Reduce cold-start by reusing running containers.
- Keep a **pool of warm workers** per function; route new invocations to an idle
  warm worker instead of creating one.
- Lifecycle policy: idle TTL, eviction under memory pressure, max reuses,
  pool-size targets. This is where the real distributed-systems tradeoffs live —
  document each choice and why.
- Cold path still works when the pool is empty/cold (tie back to Phase 2).
- **Benchmark → headline number #2:** warm-invocation latency vs. cold-start
  (the money comparison). Show the distribution, and the crossover behavior as
  traffic ramps.
- **Exit criteria:** warm latency an order of magnitude better than cold;
  before/after numbers in the README.

> At the end of Phase 4 you have a **complete, portfolio-worthy project**:
> scheduler + container execution + warm pool + three headline numbers. Per the
> working notes in CLAUDE.md, ship/write this up **before** starting Phase 5.

## Phase 5 — (Stretch) Firecracker microVM isolation *(scope item 5)*
Swap container isolation for microVMs and benchmark the difference.
- Requires the **KVM-capable Linux host** from Phase 0.
- Backend via `firecracker-go-sdk` (drive VMs directly) or
  **firecracker-containerd** (keeps your containerd integration, swaps the
  runtime) — the latter mirrors the "forking faasd" writeup in CLAUDE.md.
- Build minimal **kernel + rootfs** images for the runtime.
- Reuse the Phase 4 warm-pool logic against VMs instead of containers.
- **Benchmark → the writeup centerpiece:** containers vs. microVMs on
  cold-start, warm latency, density/overhead, and isolation guarantees.
- **Exit criteria:** same functions run on both backends; head-to-head numbers.

## Phase 6 — Deploy + writeup (runs alongside, finalized here)
- Deploy to a reachable host (the Linux box); expose the invoke API.
- Freeze the benchmark methodology and regenerate all numbers on that host.
- **README** with: architecture diagram, the quantified results table
  (cold-start, warm latency, concurrency ceiling, + container-vs-microVM if you
  did Phase 5), and the tradeoff discussion.
- Short standalone writeup/blog post — narrative tie to the CBP / HPC-scheduling
  thread (resource allocation & systems).

---

## Cross-cutting (build incrementally, not as a phase)
- **Observability:** Prometheus metrics from Phase 2 on — you can't report
  numbers you don't measure.
- **Benchmark reproducibility:** scripted, versioned, one-command re-run.
- **Executor interface:** keep subprocess / Docker / containerd / Firecracker
  behind one interface so each phase is a backend swap, not a rewrite.

## Suggested order of value
Phases 0→4 are the spine — do them in order; each is independently demoable.
Phase 5 is genuinely optional and gated on hardware. Phase 6 work (README,
numbers, deploy) should be kept current from Phase 2 onward, not saved for the end.
```
