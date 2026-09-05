# Threat Model

**Core threat:** this platform runs **untrusted user code**. Nearly every
vulnerability below flows from that. Isolation strength is the central security
property — the container-vs-microVM (Phase 5) choice is fundamentally a security
argument, not just a performance one.

---

## Critical — isolation
- **Subprocess runs as the control plane's own user/FS/network** (Phase 1).
  Malicious function code = full host access. Mitigation: containers (Phase 2) →
  microVMs (Phase 5).
- **`exec.Command` inherits the parent environment** → function sees platform
  secrets. Fix: set `proc.Env` to a minimal allowlist.
- **Container escape (future).** A shared-kernel container can be escaped via
  a kernel bug. Why Lambda uses Firecracker (separate guest kernel, small attack
  surface). Phase 5.

## High — denial of service (no resource limits)
- **Unbounded request body** (`io.ReadAll(r.Body)`) → memory exhaustion. Fix:
  `http.MaxBytesReader`.
- **No execution timeout.** Default `http.Post` client never times out; a
  function that hangs after becoming ready stalls the handler forever. Fix:
  `context` deadline on the forward call.
- **No CPU / memory / PID caps** → fork bomb or busy loop takes down the host.
  Fix: cgroup limits (containers provide these).

## High — network / SSRF
- **Full network egress from functions** → data exfiltration or attacks on
  internal services, classically the cloud metadata endpoint
  `169.254.169.254` to steal instance credentials. Fix: locked-down network
  namespace + egress allowlist.

## Medium — trust & multi-tenancy
- **No auth on the control plane (`:9000`)** — anyone who can reach it can
  invoke anything. Acceptable for local dev only.
- **Fixed `:8080`, no auth between control plane and function** — any local
  process can talk to it; with naive warm reuse a container could serve another
  tenant's leftover state/request. Fix: per-invocation port/socket; reset or
  discard the sandbox between tenants.

---

## Notes
- Revisit this file at the end of each phase; flip statuses as mitigations land.
- Isolation (Phase 2/5) is necessary but **not sufficient** — the DoS, SSRF, and
  auth items must be handled independently of the sandbox type.
