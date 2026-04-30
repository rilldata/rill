import {
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  V1FileEvent,
  V1ReconcileStatus,
  V1ResourceEvent,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { writable } from "svelte/store";

// Fake composed stream: the watcher constructs one via createSSEStream(...),
// so intercepting that factory lets us drive typed payload and connection
// events without a real transport.
class FakeSSEStream {
  public status = writable("closed");
  private readonly typedHandlers = new Map<
    string,
    Set<(arg?: unknown) => void>
  >();
  private readonly connectionHandlers = new Map<
    string,
    Set<(arg?: unknown) => void>
  >();

  public start = vi.fn();
  public close = vi.fn();
  public pause = vi.fn();
  public resumeIfPaused = vi.fn(async () => {});
  public cleanup = vi.fn();

  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.typedHandlers.has(event))
      this.typedHandlers.set(event, new Set());
    this.typedHandlers.get(event)!.add(listener);
    return () => this.typedHandlers.get(event)!.delete(listener);
  };
  public once = this.on;

  public onConnection = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.connectionHandlers.has(event)) {
      this.connectionHandlers.set(event, new Set());
    }
    this.connectionHandlers.get(event)!.add(listener);
    return () => this.connectionHandlers.get(event)!.delete(listener);
  };
  public onceConnection = this.onConnection;

  public fire(event: string, arg?: unknown) {
    this.typedHandlers.get(event)?.forEach((h) => h(arg));
  }

  public fireConnection(event: string, arg?: unknown) {
    this.connectionHandlers.get(event)?.forEach((h) => h(arg));
  }
}

const fakeStreams: FakeSSEStream[] = [];

vi.mock("@rilldata/web-common/runtime-client/sse", async (importOriginal) => {
  const actual =
    await importOriginal<
      typeof import("@rilldata/web-common/runtime-client/sse")
    >();
  return {
    ...actual,
    createSSEStream: vi.fn(() => {
      const stream = new FakeSSEStream();
      fakeStreams.push(stream);
      return stream;
    }),
  };
});

// Stub the module singletons the watcher wires into its invalidators. The
// production surface imports these directly, so replacing the modules is the
// test seam; keeps the watcher's constructor free of test-only plumbing.
vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: {
      getFileArtifact: vi.fn(() => ({
        fetchContent: vi.fn().mockResolvedValue(undefined),
      })),
      removeFile: vi.fn(),
      updateArtifacts: vi.fn(),
      deleteResource: vi.fn(),
      init: vi.fn().mockResolvedValue(undefined),
      setClient: vi.fn(),
    },
  }),
);

vi.mock("@rilldata/web-common/lib/event-bus/event-bus", () => ({
  eventBus: { emit: vi.fn() },
}));

vi.mock(
  "@rilldata/web-common/features/connectors/explorer/connector-explorer-store",
  () => ({
    connectorExplorerStore: { deleteItem: vi.fn() },
  }),
);

vi.mock("@rilldata/web-common/features/sources/sources-store", () => ({
  sourceIngestionTracker: {
    isPending: vi.fn(() => false),
    trackIngested: vi.fn(),
  },
}));

vi.mock("$app/navigation", () => ({
  invalidate: vi.fn().mockResolvedValue(undefined),
}));

import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { FileAndResourceWatcher } from "./file-and-resource-watcher";

const INSTANCE_ID = "inst-1";

function fakeQueryClient() {
  return {
    invalidateQueries: vi.fn().mockResolvedValue(undefined),
    refetchQueries: vi.fn().mockResolvedValue(undefined),
    resetQueries: vi.fn().mockResolvedValue(undefined),
    removeQueries: vi.fn(),
    fetchQuery: vi.fn().mockResolvedValue({ resources: [] }),
    getQueryData: vi.fn(() => undefined),
    setQueryData: vi.fn(),
  } as unknown as QueryClient & {
    invalidateQueries: ReturnType<typeof vi.fn>;
    refetchQueries: ReturnType<typeof vi.fn>;
    resetQueries: ReturnType<typeof vi.fn>;
    setQueryData: ReturnType<typeof vi.fn>;
  };
}

function fakeRuntimeClient() {
  return {
    getJwt: () => "tok",
    instanceId: INSTANCE_ID,
  } as never;
}

describe("FileAndResourceWatcher", () => {
  beforeEach(() => {
    fakeStreams.length = 0;
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.unstubAllEnvs();
  });

  it("routes file messages through handleFileEvent", async () => {
    const qc = fakeQueryClient();
    new FileAndResourceWatcher({
      runtimeClient: fakeRuntimeClient(),
      queryClient: qc,
      lifecycle: "none",
    });

    const stream = fakeStreams[0];
    stream.fire("file", {
      path: "/rill.yaml",
      event: V1FileEvent.FILE_EVENT_WRITE,
      isDir: false,
    });

    // Flush microtasks for async invalidators.
    await new Promise((r) => setTimeout(r, 0));

    expect(qc.invalidateQueries).toHaveBeenCalledWith({
      queryKey: getRuntimeServiceIssueDevJWTQueryKey(INSTANCE_ID),
    });
    expect(eventBus.emit).toHaveBeenCalledWith("rill-yaml-updated");
  });

  it("routes resource messages through handleResourceEvent", async () => {
    const qc = fakeQueryClient();
    new FileAndResourceWatcher({
      runtimeClient: fakeRuntimeClient(),
      queryClient: qc,
      lifecycle: "none",
    });

    const stream = fakeStreams[0];
    stream.fire("resource", {
      event: V1ResourceEvent.RESOURCE_EVENT_WRITE,
      name: { name: "mv", kind: "rill.runtime.v1.MetricsView" },
      resource: {
        meta: {
          reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
          stateVersion: "1",
        },
      },
    });
    await new Promise((r) => setTimeout(r, 0));

    // setQueryData was called with the resource key — the dispatcher ran.
    expect(qc.setQueryData).toHaveBeenCalledWith(
      getRuntimeServiceGetResourceQueryKey(INSTANCE_ID, {
        name: { name: "mv", kind: "rill.runtime.v1.MetricsView" },
      }),
      expect.objectContaining({ resource: expect.any(Object) }),
    );
  });

  it("logs resource reconcile status for playwright e2e synchronization", async () => {
    vi.stubEnv("VITE_PLAYWRIGHT_TEST", "true");
    const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

    const qc = fakeQueryClient();
    new FileAndResourceWatcher({
      runtimeClient: fakeRuntimeClient(),
      queryClient: qc,
      lifecycle: "none",
    });

    const stream = fakeStreams[0];
    stream.fire("resource", {
      event: V1ResourceEvent.RESOURCE_EVENT_WRITE,
      name: { name: "mv", kind: "rill.runtime.v1.MetricsView" },
      resource: {
        meta: {
          reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
          stateVersion: "1",
        },
      },
    });
    await new Promise((r) => setTimeout(r, 0));

    expect(logSpy).toHaveBeenCalledWith(
      "[RECONCILE_STATUS_IDLE] rill.runtime.v1.MetricsView/mv",
    );
  });

  it("on reconnect, invalidates all runtime-scoped queries", async () => {
    const qc = fakeQueryClient();
    new FileAndResourceWatcher({
      runtimeClient: fakeRuntimeClient(),
      queryClient: qc,
      lifecycle: "none",
    });

    const stream = fakeStreams[0];
    stream.fireConnection("reconnect");
    await new Promise((r) => setTimeout(r, 0));

    // The invalidateAll call uses a predicate; this asserts it was called
    // with a predicate-shaped argument.
    const predicateCall = qc.invalidateQueries.mock.calls.find(
      ([arg]) => typeof arg.predicate === "function",
    );
    expect(predicateCall).toBeDefined();
  });

  it("start(url) passes through to the underlying connection", () => {
    const qc = fakeQueryClient();
    const watcher = new FileAndResourceWatcher({
      runtimeClient: fakeRuntimeClient(),
      queryClient: qc,
      lifecycle: "none",
    });
    watcher.start("http://x/sse");

    const stream = fakeStreams[0];
    expect(stream.start).toHaveBeenCalledTimes(1);
    const [, opts] = stream.start.mock.calls[0];
    expect(opts.getJwt()).toBe("tok");
  });
});
