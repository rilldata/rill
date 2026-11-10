// web-local has no unit runner; keep its real root loader covered by web-common's CI suite.
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { isCancelledError, QueryObserver } from "@tanstack/svelte-query";
import { readable } from "svelte/store";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  V1FileEvent,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  createFileInvalidatorState,
  handleFileEvent,
  type FileInvalidatorState,
} from "@rilldata/web-common/runtime-client/invalidation/file-invalidators";
import { FileAndResourceWatcher } from "@rilldata/web-common/features/entity-management/file-and-resource-watcher";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { handleUninitializedProject } from "@rilldata/web-common/features/welcome/is-project-initialized.js";
import { load } from "../../web-local/src/routes/+layout";

const { connectionHandlers } = vi.hoisted(() => ({
  connectionHandlers: new Map<string, () => void>(),
}));
vi.mock("@rilldata/web-common/runtime-client", async (original) => ({
  ...(await original<typeof import("@rilldata/web-common/runtime-client")>()),
  runtimeServiceListFiles: vi.fn(),
}));
vi.mock("@rilldata/web-common/runtime-client/local-service", () => ({
  localServiceGetMetadata: vi.fn(async () => ({})),
}));
vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: {
      setClient: vi.fn(),
      init: vi.fn(async () => {}),
      getFileArtifact: () => ({
        fetchContent: vi.fn(async () => {}),
        isPreviewableDataFile: false,
      }),
    },
  }),
);
vi.mock("../../web-local/src/lib/runtime-client", () => ({
  getLocalRuntimeClient: () => ({ instanceId: "default" }),
}));
vi.mock("@rilldata/web-common/features/entity-management/file-io.ts", () => ({
  RuntimeFileIO: class {},
}));
vi.mock(
  "@rilldata/web-common/features/welcome/is-project-initialized.js",
  () => ({ handleUninitializedProject: vi.fn(async () => false) }),
);
vi.mock("$app/navigation", () => ({ invalidate: vi.fn() }));
vi.mock("@rilldata/web-common/runtime-client/sse", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/sse")
  >()),
  createSSEStream: () => ({
    status: readable("open"),
    on: vi.fn(),
    onConnection: (name: string, callback: () => void) =>
      connectionHandlers.set(name, callback),
    start: vi.fn(),
    close: vi.fn(),
  }),
}));

const client = { instanceId: "default" } as RuntimeClient;
const key = getRuntimeServiceListFilesQueryKey(client.instanceId, {});
const fresh: V1ListFilesResponse = {
  files: [{ path: "/rill.yaml" }, { path: "/connectors/new.yaml" }],
};

function loadRoot() {
  return load({
    url: new URL("http://localhost/"),
    depends: vi.fn(),
    untrack: (fn: () => unknown) => fn(),
    route: { id: "/" },
  } as Parameters<typeof load>[0]);
}
function observe<T>(promise: Promise<T>) {
  return promise.then(
    (value) => ({ value, error: undefined }),
    (error) => ({ value: undefined, error }),
  );
}
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: unknown) => void;
  const promise = new Promise<T>((done, fail) => {
    resolve = done;
    reject = fail;
  });
  return { promise, resolve, reject };
}

describe("local root layout ListFiles", () => {
  const states: FileInvalidatorState[] = [];
  const cleanups: (() => void)[] = [];

  beforeEach(() => {
    vi.useFakeTimers();
    queryClient.clear();
    connectionHandlers.clear();
    vi.mocked(runtimeServiceListFiles).mockReset();
  });
  afterEach(() => {
    states
      .splice(0)
      .forEach((state) => state.refetchListFilesThrottle.cancel());
    cleanups.splice(0).forEach((cleanup) => cleanup());
    queryClient.clear();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it.each(["file event", "reconnect"] as const)(
    "joins the fresh replacement after a %s supersedes its request",
    async (cause) => {
      // Cached data is what makes TanStack cancel rather than share the in-flight request.
      queryClient.setQueryData(key, { files: [] });
      const old = deferred<V1ListFilesResponse>();
      const replacement = deferred<V1ListFilesResponse>();
      vi.mocked(runtimeServiceListFiles)
        .mockImplementationOnce(() => old.promise)
        .mockImplementationOnce(() => replacement.promise);

      if (cause === "reconnect") {
        // Reconnect invalidation only refetches active queries, like the file explorer in the app.
        const observer = new QueryObserver(queryClient, {
          queryKey: key,
          staleTime: Infinity,
          queryFn: () => runtimeServiceListFiles(client, {}),
        });
        cleanups.push(observer.subscribe(() => {}));
      }
      const foreground = observe(loadRoot());
      await vi.waitFor(() =>
        expect(runtimeServiceListFiles).toHaveBeenCalledTimes(1),
      );
      const signal = vi.mocked(runtimeServiceListFiles).mock.calls[0][2]
        ?.signal;
      expect(signal).toBeInstanceOf(AbortSignal);

      if (cause === "file event") {
        const state = createFileInvalidatorState();
        states.push(state);
        await handleFileEvent(
          { path: "/connectors/new.yaml", event: V1FileEvent.FILE_EVENT_WRITE },
          queryClient,
          client,
          state,
        );
        await vi.advanceTimersByTimeAsync(101);
      } else {
        const watcher = new FileAndResourceWatcher({
          runtimeClient: client,
          queryClient,
          lifecycle: "none",
        });
        cleanups.push(() => watcher.close(true));
        connectionHandlers.get("reconnect")!();
      }
      await vi.waitFor(() =>
        expect(runtimeServiceListFiles).toHaveBeenCalledTimes(2),
      );
      expect(signal?.aborted).toBe(true);
      replacement.resolve(fresh);
      const result = await foreground;
      expect(result.error).toBeUndefined();
      expect(result.value).toMatchObject({ initialized: true });
      expect(handleUninitializedProject).not.toHaveBeenCalled();
      expect(queryClient.getQueryData(key)).toEqual(fresh);
      old.resolve({ files: [] });
      await Promise.resolve();
      expect(queryClient.getQueryData(key)).toEqual(fresh);
      expect(runtimeServiceListFiles).toHaveBeenCalledTimes(2);
    },
  );

  it("makes only one request for ordinary success", async () => {
    vi.mocked(runtimeServiceListFiles).mockResolvedValue(fresh);
    expect(await loadRoot()).toMatchObject({ initialized: true });
    expect(runtimeServiceListFiles).toHaveBeenCalledTimes(1);
  });

  it("propagates a non-cancellation error without continuing", async () => {
    const error = new Error("Permission denied");
    vi.mocked(runtimeServiceListFiles).mockRejectedValue(error);
    expect((await observe(loadRoot())).error).toBe(error);
    expect(runtimeServiceListFiles).toHaveBeenCalledTimes(1);
  });

  it("propagates an error from the replacement without another request", async () => {
    queryClient.setQueryData(key, { files: [] });
    const old = deferred<V1ListFilesResponse>();
    const replacement = deferred<V1ListFilesResponse>();
    const error = new Error("Permission denied");
    vi.mocked(runtimeServiceListFiles)
      .mockImplementationOnce(() => old.promise)
      .mockImplementationOnce(() => replacement.promise);
    const result = observe(loadRoot());
    await vi.waitFor(() =>
      expect(runtimeServiceListFiles).toHaveBeenCalledTimes(1),
    );
    const background = queryClient.refetchQueries({ queryKey: key });
    await vi.waitFor(() =>
      expect(runtimeServiceListFiles).toHaveBeenCalledTimes(2),
    );
    replacement.reject(error);
    await background;
    expect((await result).error).toBe(error);
    expect(runtimeServiceListFiles).toHaveBeenCalledTimes(2);
    old.resolve({ files: [] });
  });

  it("rethrows the third cancellation unchanged instead of continuing indefinitely", async () => {
    const requests: ReturnType<typeof deferred<V1ListFilesResponse>>[] = [];
    vi.mocked(runtimeServiceListFiles).mockImplementation(() => {
      const request = deferred<V1ListFilesResponse>();
      requests.push(request);
      return request.promise;
    });
    const result = observe(loadRoot());
    for (let request = 1; request <= 2; request++) {
      await vi.waitFor(() =>
        expect(runtimeServiceListFiles).toHaveBeenCalledTimes(request),
      );
      await queryClient.cancelQueries({ queryKey: key });
    }
    await vi.waitFor(() =>
      expect(runtimeServiceListFiles).toHaveBeenCalledTimes(3),
    );
    // A second awaiter sees the identical error that the loader must propagate.
    const lastAwaiter = observe(queryClient.fetchQuery({ queryKey: key }));
    await queryClient.cancelQueries({ queryKey: key });
    const final = await result;
    expect(isCancelledError(final.error)).toBe(true);
    expect(final.error).toBe((await lastAwaiter).error);
    expect(runtimeServiceListFiles).toHaveBeenCalledTimes(3);
    requests.forEach((request) => request.resolve({ files: [] }));
  });
});
