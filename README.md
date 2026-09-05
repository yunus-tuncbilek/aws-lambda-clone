# faas — a Function-as-a-Service platform (Lambda-style), from scratch

A minimal FaaS platform in Go: send an HTTP event, get a fresh isolated container
spun up to handle it, get the response back, container torn down. Built to go
**deep** on the hard parts of serverless infrastructure — isolation, scheduling,
and cold-start latency — rather than wide on feature parity with AWS Lambda.

> **Status: work in progress (Phase 2 of 6 complete).** Container-per-invocation
> execution works end to end with real cold-start numbers. See the
> [Roadmap](#roadmap) for exactly what's done and what's coming.

## Why this project

Serverless platforms hide an enormous amount of systems engineering behind a
single `invoke` call. The goal here is to rebuild that machinery to understand
the tradeoffs first-hand — containers vs. microVMs, warm-pool eviction policy,
admission control under load — at a level I can defend in detail, backed by
measured latency numbers.

## How it works today

```
  client                control plane (:9000)              function container
  ──────                ─────────────────────              ──────────────────
  POST /invoke/hello ─────────►  look up image
                                 create container  ──────►  cold-start
                                 (containerd task)          server on :8080
                                 wait for :8080 ready
                                 forward event   ──────────► POST /
                                                 ◄────────── response body
                        ◄─────  return response
                                 kill + delete container
```

- **Control plane** ([`cmd/controlplane`](cmd/controlplane)) — an HTTP server that
  receives invocations and drives the executor.
- **Executor** ([`executor.go`](cmd/controlplane/executor.go)) — talks to
  **containerd** directly: pulls the image reference, creates a container with a
  fresh writable snapshot and OCI runtime spec, starts the task, forwards the
  event to the function's port, and tears everything down afterward.
- **Function contract** ([`docs/function-contract.md`](docs/function-contract.md))
  — a function is a container image running an HTTP server on `:8080`; the platform
  `POST`s the event and returns the response body. This "server-inside-container"
  model is what makes warm-container reuse (Phase 4) possible.
- **Example function** ([`functions/hello`](functions/hello)) — a tiny Go echo
  handler, built into an ~8 MB `scratch` image.

## Results so far

Measured inside a Colima VM (Apple Virtualization.Framework, arm64), containerd
2.3.4, `hello` on a `scratch` image. 

| Backend | Isolation | Cold-start latency |
|---|---|---|
| Subprocess (Phase 1) | none | ~15 ms |
| Container per invocation (Phase 2) | namespaces + cgroups | **~65 ms** (51–82) |

Conclusion: **Container isolation costs ~50 ms per cold start** over a bare subprocess.

Full methodology in
[`docs/benchmarks.md`](docs/benchmarks.md).

## Running it

Requires a Linux host with containerd (on macOS: [Colima](https://github.com/abiosoft/colima)
provides one inside a VM). From inside the VM:

```sh
# build the example function image (nerdctl/containerd)
nerdctl build -t hello:latest -f functions/hello/Dockerfile .

# run the control plane
go run ./cmd/controlplane

# invoke it
curl -X POST localhost:9000/invoke/hello -d 'world'
# -> hello, world
```

## Design docs

- [Function contract](docs/function-contract.md) — what "a function" is
- [Threat model](docs/threat-model.md) — this platform runs untrusted code; why
  isolation is the central security property
- [Benchmarks](docs/benchmarks.md) — methodology and numbers

## Reference implementations studied

faasd / OpenFaaS, firecracker-containerd, Knative, OpenWhisk, and AWS's
Firecracker microVM design.
