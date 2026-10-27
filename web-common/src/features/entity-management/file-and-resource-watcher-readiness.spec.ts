import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";
import { QueryClient } from "@tanstack/svelte-query";
import { runtimeServiceGetResource } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  ConnectionStatus,
  FileAndResourceWatcher,
} from "./file-and-resource-watcher";

// Keep the real watcher, subscriber and reconnect state machine. Only the
// network transport is replaced, so it can reject opens during a restart.
let unavailableUntil = 0;
class RestartingTransport {
  private handlers = new Map<string, Set<(arg?: unknown) => void>>();
  public start = vi.fn(async () => {
    await Promise.resolve();
    if (Date.now() < unavailableUntil) {
      this.fire("error", new Error("controller is closed"));
    } else {
      this.fire("open");
    }
  });
  public stop = vi.fn();
  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(listener);
    return () => this.handlers.get(event)!.delete(listener);
  };
  public once = this.on;
  public fire(event: string, arg?: unknown) {
    this.handlers.get(event)?.forEach((handler) => handler(arg));
  }
}
const transports: RestartingTransport[] = [];

vi.mock("@rilldata/web-common/runtime-client/sse/sse-fetch-client", () => ({
  SSEFetchClient: class {
    constructor() {
      const transport = new RestartingTransport();
      transports.push(transport);
      return transport;
    }
  },
  SSEHttpError: class extends Error {},
}));
vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@rilldata/web-common/runtime-client")
  >()),
  runtimeServiceGetResource: vi.fn(),
}));
vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: { init: vi.fn().mockResolvedValue(undefined) },
  }),
);
vi.mock("$app/navigation", () => ({ invalidate: vi.fn() }));

describe("watcher readiness during a controller restart", () => {
  let queryClient: QueryClient;
  let watcher: FileAndResourceWatcher;

  beforeEach(() => {
    vi.useFakeTimers();
    transports.length = 0;
    unavailableUntil = 0;
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    vi.mocked(runtimeServiceGetResource).mockImplementation(async () => {
      if (Date.now() < unavailableUntil) {
        throw { code: 14, message: "controller is closed" };
      }
      return { resource: { meta: { version: "1" } } };
    });
    watcher = new FileAndResourceWatcher({
      runtimeClient: {
        instanceId: "restart-test",
        getJwt: () => "jwt",
      } as RuntimeClient,
      queryClient,
      lifecycle: "none",
    });
  });

  afterEach(() => {
    watcher.close(true);
    queryClient.clear();
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  // With the original code, quick restarts pass and a 7-second restart
  // consumes all three retries, leaving the UI on its fatal error page.
  it.each([0, 3_000, 7_000])(
    "reopens after a %i ms restart",
    async (restartDuration) => {
      watcher.start("http://runtime/sse");
      await vi.advanceTimersByTimeAsync(5_000);
      expect(get(watcher.status)).toBe(ConnectionStatus.OPEN);

      unavailableUntil = Date.now() + restartDuration;
      transports[0].fire("close");
      await vi.advanceTimersByTimeAsync(8_000);

      expect(get(watcher.status)).toBe(ConnectionStatus.OPEN);
    },
  );

  it("still closes when the controller never becomes available", async () => {
    watcher.start("http://runtime/sse");
    await vi.advanceTimersByTimeAsync(5_000);
    unavailableUntil = Infinity;
    transports[0].fire("close");
    await vi.advanceTimersByTimeAsync(70_000);

    expect(get(watcher.status)).toBe(ConnectionStatus.CLOSED);
  });

  it("does not reconnect after the watcher is closed during a restart", async () => {
    watcher.start("http://runtime/sse");
    await vi.advanceTimersByTimeAsync(5_000);
    unavailableUntil = Date.now() + 7_000;
    transports[0].fire("close");
    await vi.advanceTimersByTimeAsync(1_000);
    watcher.close();
    const attemptsAtClose = transports[0].start.mock.calls.length;
    await vi.advanceTimersByTimeAsync(10_000);

    expect(get(watcher.status)).toBe(ConnectionStatus.CLOSED);
    expect(transports[0].start).toHaveBeenCalledTimes(attemptsAtClose);
  });
});
