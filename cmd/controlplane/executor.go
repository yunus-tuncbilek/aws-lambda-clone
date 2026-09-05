package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	containerdSocket = "/run/containerd/containerd.sock"
	containerdNS     = "default"     // namespace where nerdctl built our image
	functionAddr     = "localhost:8080"
)

// executor runs functions as containerd containers, one per invocation.
type executor struct {
	client *containerd.Client
}

func newExecutor() (*executor, error) {
	client, err := containerd.New(containerdSocket)
	if err != nil {
		return nil, fmt.Errorf("connect to containerd: %w", err)
	}
	return &executor{client: client}, nil
}

// run cold-starts a container from imageRef, forwards the event to its :8080,
// returns the response, then tears the container down. Mirrors the Phase 1
// subprocess flow — only the "how we start it" changed.
func (e *executor) run(imageRef string, event []byte) (body []byte, status int, err error) {
	// containerd scopes everything (images, containers) by namespace.
	ctx := namespaces.WithNamespace(context.Background(), containerdNS)

	// The image was already pulled/built by nerdctl; just look it up.
	image, err := e.client.GetImage(ctx, imageRef)
	if err != nil {
		return nil, 0, fmt.Errorf("get image: %w", err)
	}

	id := fmt.Sprintf("fn-%d", time.Now().UnixNano())

	// Container = metadata only: which image, a fresh writable snapshot
	// (its private scratch layer), and the OCI runtime spec. Sharing the host
	// (VM) network namespace makes the container's :8080 reachable on localhost.
	container, err := e.client.NewContainer(ctx, id,
		containerd.WithNewSnapshot(id+"-snap", image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithHostNamespace(specs.NetworkNamespace),
		),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("new container: %w", err)
	}
	defer container.Delete(ctx, containerd.WithSnapshotCleanup)

	// Task = the actual running process created from that container.
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return nil, 0, fmt.Errorf("new task: %w", err)
	}
	// Get the exit channel BEFORE starting, then define teardown: kill the
	// process, wait for it to actually exit, then delete the task.
	exitCh, err := task.Wait(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("wait task: %w", err)
	}
	defer func() {
		task.Kill(ctx, syscall.SIGKILL)
		select {
		case <-exitCh:
		case <-time.After(2 * time.Second):
		}
		task.Delete(ctx)
	}()

	if err := task.Start(ctx); err != nil {
		return nil, 0, fmt.Errorf("start task: %w", err)
	}

	// Everything below is identical to Phase 1: wait for readiness, forward.
	if err := waitReady(functionAddr, 5*time.Second); err != nil {
		return nil, 0, fmt.Errorf("function not ready: %w", err)
	}

	resp, err := http.Post("http://"+functionAddr+"/", "application/octet-stream", bytes.NewReader(event))
	if err != nil {
		return nil, 0, fmt.Errorf("forward event: %w", err)
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}
