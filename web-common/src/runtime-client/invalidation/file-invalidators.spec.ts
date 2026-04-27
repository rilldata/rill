import { Throttler } from "@rilldata/web-common/lib/throttler";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGitStatusQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  getRuntimeServiceListFilesQueryKey,
  V1FileEvent,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { fetchContent, removeFile } = vi.hoisted(() => ({
  fetchContent: vi.fn().mockResolvedValue(undefined),
  removeFile: vi.fn(),
}));

vi.mock("$app/navigation", () => ({
  invalidate: vi.fn().mockResolvedValue(undefined),
}));

vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: {
      getFileArtifact: vi.fn(() => ({ fetchContent })),
      removeFile,
    },
  }),
);

vi.mock("@rilldata/web-common/lib/event-bus/event-bus", () => ({
  eventBus: { emit: vi.fn() },
}));

import { invalidate } from "$app/navigation";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  handleFileEvent,
  type FileInvalidatorState,
} from "./file-invalidators";

const INSTANCE_ID = "inst-1";
const fakeRuntimeClient = { instanceId: INSTANCE_ID } as RuntimeClient;

function fakeQueryClient() {
  return {
    invalidateQueries: vi.fn(),
    refetchQueries: vi.fn(),
    resetQueries: vi.fn(),
  } as unknown as QueryClient & {
    invalidateQueries: ReturnType<typeof vi.fn>;
    refetchQueries: ReturnType<typeof vi.fn>;
    resetQueries: ReturnType<typeof vi.fn>;
  };
}

function makeState(): FileInvalidatorState & {
  refetchListFilesThrottle: Throttler;
} {
  return {
    seenFiles: new Set<string>(),
    refetchListFilesThrottle: new Throttler(0, 0),
  };
}

describe("handleFileEvent", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("short-circuits on .db paths without touching the query client", async () => {
    const qc = fakeQueryClient();
    const state = makeState();

    await handleFileEvent(
      {
        path: "/data/foo.db",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );

    expect(qc.invalidateQueries).not.toHaveBeenCalled();
    expect(qc.refetchQueries).not.toHaveBeenCalled();
    expect(fetchContent).not.toHaveBeenCalled();
  });

  it("write on /rill.yaml invalidates the dev JWT key + reruns app:init + emits rill-yaml-updated", async () => {
    const qc = fakeQueryClient();
    const state = makeState();

    await handleFileEvent(
      {
        path: "/rill.yaml",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );

    expect(qc.invalidateQueries).toHaveBeenCalledWith({
      queryKey: getRuntimeServiceIssueDevJWTQueryKey(INSTANCE_ID),
    });
    expect(invalidate).toHaveBeenCalledWith("app:init");
    expect(eventBus.emit).toHaveBeenCalledWith("rill-yaml-updated");
    expect(state.seenFiles.has("/rill.yaml")).toBe(true);
  });

  it("delete on /rill.yaml reruns app:init but does not invalidate the dev JWT key", async () => {
    const qc = fakeQueryClient();
    const state = makeState();
    state.seenFiles.add("/rill.yaml");

    await handleFileEvent(
      {
        path: "/rill.yaml",
        event: V1FileEvent.FILE_EVENT_DELETE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );

    expect(invalidate).toHaveBeenCalledWith("app:init");

    const devJwtKey = getRuntimeServiceIssueDevJWTQueryKey(INSTANCE_ID);
    const devJwtHit = qc.invalidateQueries.mock.calls.some(
      ([arg]) => arg.queryKey === devJwtKey,
    );
    expect(devJwtHit).toBe(false);
    expect(state.seenFiles.has("/rill.yaml")).toBe(false);
  });

  it("write on a new file triggers a throttled listFiles refetch", async () => {
    const qc = fakeQueryClient();
    const state = makeState();

    await handleFileEvent(
      {
        path: "/models/foo.sql",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );
    await vi.advanceTimersByTimeAsync(10);

    expect(qc.refetchQueries).toHaveBeenCalledWith({
      queryKey: getRuntimeServiceListFilesQueryKey(INSTANCE_ID),
    });
    expect(state.seenFiles.has("/models/foo.sql")).toBe(true);
  });

  it("write on an already-seen file does not trigger a listFiles refetch", async () => {
    const qc = fakeQueryClient();
    const state = makeState();
    state.seenFiles.add("/models/foo.sql");

    await handleFileEvent(
      {
        path: "/models/foo.sql",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );
    await vi.advanceTimersByTimeAsync(10);

    expect(qc.refetchQueries).not.toHaveBeenCalled();
  });

  it("delete always triggers a listFiles refetch", async () => {
    const qc = fakeQueryClient();
    const state = makeState();
    state.seenFiles.add("/models/foo.sql");

    await handleFileEvent(
      {
        path: "/models/foo.sql",
        event: V1FileEvent.FILE_EVENT_DELETE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );
    await vi.advanceTimersByTimeAsync(10);

    expect(qc.refetchQueries).toHaveBeenCalledWith({
      queryKey: getRuntimeServiceListFilesQueryKey(INSTANCE_ID),
    });
    expect(qc.resetQueries).toHaveBeenCalledWith({
      queryKey: getRuntimeServiceGetFileQueryKey(INSTANCE_ID, {
        path: "/models/foo.sql",
      }),
    });
    expect(removeFile).toHaveBeenCalledWith("/models/foo.sql");
  });

  it("invalidates git status on a non-dir file event", async () => {
    const qc = fakeQueryClient();
    const state = makeState();

    await handleFileEvent(
      {
        path: "/models/foo.sql",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: false,
      },
      qc,
      fakeRuntimeClient,
      state,
    );

    const gitStatusKey = getRuntimeServiceGitStatusQueryKey(INSTANCE_ID, {});
    const gitHit = qc.invalidateQueries.mock.calls.some(
      ([arg]) =>
        Array.isArray(arg.queryKey) &&
        JSON.stringify(arg.queryKey) === JSON.stringify(gitStatusKey),
    );
    expect(gitHit).toBe(true);
  });

  it("does not invalidate git status on a dir event", async () => {
    const qc = fakeQueryClient();
    const state = makeState();

    await handleFileEvent(
      {
        path: "/models",
        event: V1FileEvent.FILE_EVENT_WRITE,
        isDir: true,
      },
      qc,
      fakeRuntimeClient,
      state,
    );

    expect(qc.invalidateQueries).not.toHaveBeenCalled();
  });
});
