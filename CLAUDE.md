# Project: FaaS Platform (AWS Lambda Clone)

## Working style — LEARNING-FIRST (read this before doing anything)
The point of this project is for ME to learn, not just to produce a working
system. Speed is not the goal; comprehension is. Therefore:

- **Check in before introducing anything new.** Before using a new tool,
  framework, library, or programming language — or before implementing any
  major or technically-complex feature — STOP and confirm I understand what
  you're about to do. Do not just proceed.
- **Ask comprehension questions first.** For example: "Do you have experience
  with the Go programming language?", "Are you familiar with this method for
  handling concurrency?", "Do you know how containers isolate processes?" Gauge
  my level, then teach to it.
- **Explain before you build.** Walk me through the concept, the tradeoffs, and
  why this approach — in plain terms — and let me respond before writing code.
- **Prefer teaching over hand-waving.** If something is genuinely complex, slow
  down and break it down rather than abstracting it away.
- When in doubt about whether I'll understand a step, assume I want to be asked.
- Keep your answers as short as possible without jeopardizing the other goals.

## Purpose
Personal project to (1) deepen infrastructure/systems knowledge and (2) build a
portfolio piece distinct from prior CRUD-app work. Not aiming for AWS-scale
parity — the goal is to go deep on the hard distributed-systems problems, not
wide on Lambda's full feature surface.

## Scope — build these deeply
1. HTTP-triggered function execution (event -> run -> response)
2. Scheduler/queue for concurrent request handling
3. Container-based isolation, spun up per invocation
4. Warm-pool reuse to reduce cold-start latency
5. (Stretch) VM-level isolation via Firecracker instead of plain containers

## Explicitly out of scope
- API Gateway / IAM / DynamoDB-trigger equivalents
- Multi-region, multi-tenant billing, or other AWS-parity features
- Anything that trades depth on the above 4-5 items for breadth

## Reference material
- Tomasz Janczuk's serverless platform blueprint (ex-Auth0 Extend) — architecture lessons from a real production FaaS
- AWS Firecracker docs/announcement — microVM isolation model, why Lambda is architected as it is
- OpenFaaS / faasd / Knative / OpenWhisk source — mature reference implementations worth reading, not just tutorials
- Writeup: forking faasd to swap in firecracker-containerd (container vs. microVM isolation benchmarking)

## What makes this portfolio-worthy (bar to hit)
- Depth over breadth: be able to explain tradeoffs (containers vs. microVMs,
  warm-pool strategy, scheduler design) in interview-level detail
- Quantified results in the README: cold-start latency, warm-invocation
  latency, concurrent-request ceiling before failure — numbers, not just
  "it works"
- Deployed somewhere reachable + a short writeup (not just a repo)
- Narrative fit: complements existing CBP (LP solver / Gurobi / HPC scheduling)
  experience — same "resource allocation / systems" thread, different project

## Working notes
- Prefer shipping a smaller, well-documented version (scheduler + container
  execution + real benchmarks) sooner over a half-finished microVM version later