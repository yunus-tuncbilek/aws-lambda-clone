# ---- build stage: has the full Go toolchain ----
FROM golang:1.27 AS build
WORKDIR /src

# Copy the module definition first (its own layer) so dependency resolution is
# cached and only re-runs when go.mod changes.
COPY go.mod ./
# Copy just this function's source.
COPY functions/hello/ ./functions/hello/

# Build a fully static Linux binary:
#   CGO_ENABLED=0 -> no libc dependency, so it runs on an empty base image
#   GOOS=linux    -> target the container's OS, not your Mac
RUN CGO_ENABLED=0 GOOS=linux go build -o /hello ./functions/hello

# ---- run stage: nothing but the binary ----
FROM scratch
COPY --from=build /hello /hello
EXPOSE 8080
ENTRYPOINT ["/hello"]
