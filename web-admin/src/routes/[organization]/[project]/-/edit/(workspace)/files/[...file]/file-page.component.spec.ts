import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
import type { FileIO } from "@rilldata/web-common/features/entity-management/file-io";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";

const client = vi.hoisted(
  () => ({ host: "http://runtime.test", instanceId: "default" }) as const,
);

vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => client,
}));
// The real workspace needs a runtime; the stub records the editor content it mounted with.
vi.mock(
  "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte",
  async () => ({
    default: (await import("./__fixtures__/MountedContent.svelte")).default,
  }),
);

import Page from "./+page.svelte";

// Answers reads only when the test says so, like a runtime that is slow to respond.
class HeldFileIO implements FileIO {
  private readonly pending = new Map<
    string,
    (blob: string | undefined) => void
  >();

  updateClient() {}

  read(path: string) {
    return new Promise<string | undefined>((resolve) =>
      this.pending.set(path, resolve),
    );
  }

  async write() {}

  respond(path: string, blob: string | undefined) {
    this.pending.get(path)?.(blob);
    this.pending.delete(path);
  }
}

function newArtifact(io: HeldFileIO, path: string) {
  return new FileArtifact(client as unknown as RuntimeClient, path, io);
}

function pageProps(fileArtifact: FileArtifact) {
  return { data: { fileArtifact } } as never;
}

afterEach(() => cleanup());

describe("cloud editor file page", () => {
  it("waits for the file content before mounting the workspace", async () => {
    const io = new HeldFileIO();
    const model = newArtifact(io, "/models/orders.yaml");

    render(Page, pageProps(model));

    expect(screen.queryByTestId("workspace")).toBeNull();

    io.respond(model.path, "type: model\nsql: select 1\n");

    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe(
        "type: model\nsql: select 1\n",
      ),
    );
  });

  it("shows a file whose content is loaded at once and waits for one that is not", async () => {
    const io = new HeldFileIO();
    const loaded = newArtifact(io, "/models/loaded.yaml");
    const notLoaded = newArtifact(io, "/models/not_loaded.yaml");
    const fetching = loaded.fetchContent();
    io.respond(loaded.path, "loaded");
    await fetching;

    const { rerender } = render(Page, pageProps(loaded));
    expect(screen.getByTestId("workspace").textContent).toBe("loaded");

    await rerender(pageProps(notLoaded));
    expect(screen.queryByTestId("workspace")).toBeNull();

    io.respond(notLoaded.path, "not loaded");
    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe("not loaded"),
    );
  });

  it("shows an error instead of an empty editor when the fetch fails, and Retry loads the file", async () => {
    const io = new HeldFileIO();
    const model = newArtifact(io, "/models/orders.yaml");

    render(Page, pageProps(model));
    io.respond(model.path, undefined);

    const retry = await screen.findByRole("button", { name: m.common_retry() });
    expect(screen.queryByTestId("workspace")).toBeNull();

    await fireEvent.click(retry);
    expect(screen.queryByRole("alert")).toBeNull();
    io.respond(model.path, "type: model\n");

    await waitFor(() =>
      expect(screen.getByTestId("workspace").textContent).toBe("type: model\n"),
    );
  });

  it("ignores a slower fetch for the previous file", async () => {
    const io = new HeldFileIO();
    const slow = newArtifact(io, "/models/slow.yaml");
    const failing = newArtifact(io, "/models/failing.yaml");

    const { rerender } = render(Page, pageProps(slow));
    await rerender(pageProps(failing));
    io.respond(failing.path, undefined);
    await screen.findByRole("alert");

    io.respond(slow.path, undefined);
    // Let the slow fetch's completion and any re-render run.
    await new Promise((resolve) => setTimeout(resolve));

    expect(screen.getByRole("alert")).toBeTruthy();
    expect(screen.queryByTestId("workspace")).toBeNull();
  });
});
