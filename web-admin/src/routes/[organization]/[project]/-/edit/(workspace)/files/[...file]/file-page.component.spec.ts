import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import { RuntimeFileIO } from "@rilldata/web-common/features/entity-management/file-io";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  GetFileResponse,
  type GetFileRequest,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
import { getRuntimeServiceGetFileQueryKey } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

const current = vi.hoisted(() => ({ runtime: undefined as unknown }));

vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => current.runtime,
}));
// The real workspace needs a runtime; the stub records the editor content it mounted with.
vi.mock(
  "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte",
  async () => ({
    default: (await import("./__fixtures__/MountedContent.svelte")).default,
  }),
);

import Page from "./+page.svelte";

type PendingRead = {
  resolve: (response: GetFileResponse) => void;
  reject: (error: Error) => void;
  signal?: AbortSignal;
};

// A runtime that answers GetFile only when the test says so, like one that is slow to respond.
class HeldRuntime {
  readonly host = "http://runtime.test";
  readonly instanceId = "default";
  private readonly pending = new Map<string, PendingRead>();

  readonly runtimeService = {
    getFile: (request: GetFileRequest, options?: { signal?: AbortSignal }) =>
      new Promise<GetFileResponse>((resolve, reject) => {
        const signal = options?.signal;
        this.pending.set(request.path, { resolve, reject, signal });
        signal?.addEventListener("abort", () => reject(signal.reason));
      }),
  };

  respond(path: string, blob: string) {
    this.pending.get(path)?.resolve(new GetFileResponse({ blob }));
  }

  fail(path: string) {
    this.pending.get(path)?.reject(new Error("runtime restarting"));
  }

  aborted(path: string) {
    return !!this.pending.get(path)?.signal?.aborted;
  }
}

function setup() {
  const runtime = new HeldRuntime();
  current.runtime = runtime;
  const client = runtime as unknown as RuntimeClient;
  const io = new RuntimeFileIO();
  io.updateClient(client);
  const newArtifact = (path: string) => new FileArtifact(client, path, io);
  return { runtime, newArtifact };
}

function pageProps(fileArtifact: FileArtifact) {
  return { data: { fileArtifact } } as never;
}

afterEach(() => {
  cleanup();
  queryClient.clear();
});

describe("cloud editor file page", () => {
  it("waits for the file content before mounting the workspace", async () => {
    const { runtime, newArtifact } = setup();
    const model = newArtifact("/models/orders.yaml");

    render(Page, pageProps(model));

    expect(screen.queryByTestId("workspace")).toBeNull();

    runtime.respond(model.path, "type: model\nsql: select 1\n");

    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe(
        "type: model\nsql: select 1\n",
      ),
    );
  });

  it("shows a file whose content is loaded at once and waits for one that is not", async () => {
    const { runtime, newArtifact } = setup();
    const loaded = newArtifact("/models/loaded.yaml");
    const notLoaded = newArtifact("/models/not_loaded.yaml");
    const fetching = loaded.fetchContent();
    runtime.respond(loaded.path, "loaded");
    await fetching;

    const { rerender } = render(Page, pageProps(loaded));
    expect(screen.getByTestId("workspace").textContent).toBe("loaded");

    await rerender(pageProps(notLoaded));
    expect(screen.queryByTestId("workspace")).toBeNull();

    runtime.respond(notLoaded.path, "not loaded");
    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe("not loaded"),
    );
  });

  it("shows an error instead of an empty editor when the fetch fails, and Retry loads the file", async () => {
    const { runtime, newArtifact } = setup();
    const model = newArtifact("/models/orders.yaml");

    render(Page, pageProps(model));
    runtime.fail(model.path);

    const retry = await screen.findByRole("button", { name: m.common_retry() });
    expect(screen.queryByTestId("workspace")).toBeNull();

    await fireEvent.click(retry);
    expect(screen.queryByRole("alert")).toBeNull();
    runtime.respond(model.path, "type: model\n");

    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe("type: model\n"),
    );
  });

  it("cancels the previous file's pending read when switching files", async () => {
    const { runtime, newArtifact } = setup();
    const slow = newArtifact("/models/slow.yaml");
    const failing = newArtifact("/models/failing.yaml");

    const { rerender } = render(Page, pageProps(slow));
    await rerender(pageProps(failing));
    expect(runtime.aborted(slow.path)).toBe(true);

    runtime.fail(failing.path);
    await screen.findByRole("alert");

    // A late answer for the cancelled read neither lands in the cache nor hides the error.
    runtime.respond(slow.path, "late");
    await new Promise((resolve) => setTimeout(resolve));

    expect(
      queryClient.getQueryData(
        getRuntimeServiceGetFileQueryKey(runtime.instanceId, {
          path: slow.path,
        }),
      ),
    ).toBeUndefined();
    expect(screen.getByRole("alert")).toBeTruthy();
  });
});
