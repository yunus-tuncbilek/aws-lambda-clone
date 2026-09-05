# Function Contract (v1)

A **function** is a container image that runs an HTTP server. The platform
invokes it by sending the event as an HTTP request; the response body is the
function's result. This "server-inside-container" model (like AWS Lambda /
OpenFaaS of-watchdog) is what enables warm-container reuse.

## The contract
- The function process listens on **port 8080** (`0.0.0.0:8080`).
- The platform sends **`POST /`** with the event as the request body.
- The function returns the result as the **response body**; HTTP status signals
  success (2xx) or error.
- Startup: the function must be ready to accept connections promptly after boot
  (cold start = time until this port answers).

## Not yet decided (revisit later)
- Health/readiness endpoint (e.g. `GET /health`).
- Request/response content type + envelope format.
- Timeout + max payload size.
