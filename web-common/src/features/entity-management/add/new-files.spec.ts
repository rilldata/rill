import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ResourceKind } from "../resource-selectors.ts";
import {
  createResourceAndNavigate,
  getPathForNewResourceFile,
} from "./new-files.ts";
import {
  runtimeServiceGetResource,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
import { fileArtifacts } from "../file-artifacts.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

vi.mock("@rilldata/web-common/runtime-client", async (original) => ({
  ...(await original<typeof import("@rilldata/web-common/runtime-client")>()),
  runtimeServiceGetResource: vi.fn(),
  runtimeServicePutFile: vi.fn(),
}));
vi.mock(
  "@rilldata/web-common/layout/navigation/editor-routing",
  async (original) => ({
    ...(await original<
      typeof import("@rilldata/web-common/layout/navigation/editor-routing")
    >()),
    navigateToFile: vi.fn(async () => {}),
  }),
);
vi.mock("@rilldata/web-common/features/file-explorer/telemetry.ts", () => ({
  getScreenNameFromPage: () => "editor",
}));
const { fireEvent } = vi.hoisted(() => ({ fireEvent: vi.fn(async () => {}) }));
vi.mock("@rilldata/web-common/metrics/initMetrics.ts", () => ({
  behaviourEvent: { fireSourceTriggerEvent: fireEvent },
}));

describe("getPathForNewResourceFile", () => {
  it("uses the current directory when creating a component", () => {
    expect(
      getPathForNewResourceFile(ResourceKind.Component, undefined, "Something"),
    ).toMatch(/^Something\/component/);
  });
});

describe("createResourceAndNavigate canvas creation", () => {
  const client = { instanceId: "default" } as RuntimeClient;
  const getResource = vi.mocked(runtimeServiceGetResource);
  const navigate = vi.mocked(navigateToFile);
  const notFound = { code: 5 };
  const resource = {
    meta: { name: { kind: ResourceKind.Canvas, name: "canvas" } },
  };

  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
    getResource.mockReset().mockResolvedValue({ resource });
    vi.mocked(runtimeServicePutFile).mockReset().mockResolvedValue({});
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  async function assertNavigated(
    pending: Promise<void>,
    path = "/dashboards/canvas.yaml",
  ) {
    await pending;
    expect(navigate).toHaveBeenCalledExactlyOnceWith(path);
    expect(fireEvent).toHaveBeenCalledTimes(1);
    expect(navigate.mock.invocationCallOrder[0]).toBeLessThan(
      fireEvent.mock.invocationCallOrder[0],
    );
    expect(vi.getTimerCount()).toBe(0);
  }

  it("waits for the canvas to exist before navigating", async () => {
    getResource.mockRejectedValueOnce(notFound);
    const pending = createResourceAndNavigate(client, ResourceKind.Canvas);
    await vi.advanceTimersByTimeAsync(0);
    expect(navigate).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(49);
    expect(navigate).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(getResource).toHaveBeenCalledTimes(2);
    expect(getResource.mock.calls[0][1]).toEqual({
      name: { kind: ResourceKind.Canvas, name: "canvas" },
    });
    await assertNavigated(pending);
  });

  it.each([undefined, "configuration error"])(
    "opens an existing canvas even with reconcile error %s",
    async (reconcileError) => {
      getResource.mockResolvedValue({
        resource: { ...resource, meta: { ...resource.meta, reconcileError } },
      });
      await assertNavigated(
        createResourceAndNavigate(client, ResourceKind.Canvas),
      );
      expect(getResource).toHaveBeenCalledTimes(1);
    },
  );

  it("navigates at the bound if the canvas remains missing", async () => {
    getResource.mockRejectedValue(notFound);
    const pending = createResourceAndNavigate(client, ResourceKind.Canvas);
    await vi.advanceTimersByTimeAsync(999);
    expect(navigate).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    await assertNavigated(pending);
    const calls = getResource.mock.calls.length;
    await vi.advanceTimersByTimeAsync(2_000);
    expect(getResource).toHaveBeenCalledTimes(calls);
  });

  it("aborts a pending lookup at the bound and still navigates", async () => {
    let signal: AbortSignal | undefined;
    getResource.mockImplementation(
      (_client, _request, options) =>
        new Promise((_resolve, reject) => {
          signal = options?.signal;
          signal?.addEventListener(
            "abort",
            () => reject(new Error("aborted")),
            { once: true },
          );
        }),
    );
    const pending = createResourceAndNavigate(client, ResourceKind.Canvas);
    await vi.advanceTimersByTimeAsync(999);
    expect(navigate).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(signal?.aborted).toBe(true);
    await assertNavigated(pending);
    expect(getResource).toHaveBeenCalledTimes(1);
  });

  it("retries a controller restart but opens immediately on an unexpected lookup error", async () => {
    getResource
      .mockRejectedValueOnce({ code: 14, message: "controller is closed" })
      .mockRejectedValueOnce(new Error("offline"));
    const pending = createResourceAndNavigate(client, ResourceKind.Canvas);
    await vi.advanceTimersByTimeAsync(50);
    await assertNavigated(pending);
    expect(getResource).toHaveBeenCalledTimes(2);
  });

  it("does not probe non-canvas creation", async () => {
    await assertNavigated(
      createResourceAndNavigate(client, ResourceKind.Model),
      "/models/model.sql",
    );
    expect(getResource).not.toHaveBeenCalled();
  });

  it("opens immediately on an unexpected lookup error", async () => {
    getResource.mockRejectedValueOnce(new Error("offline"));
    await assertNavigated(
      createResourceAndNavigate(client, ResourceKind.Canvas),
    );
    expect(getResource).toHaveBeenCalledTimes(1);
  });

  it("uses the resource base name for a suffixed canvas in a custom folder", async () => {
    vi.spyOn(fileArtifacts, "getNamesForKind").mockReturnValue(["canvas"]);
    await assertNavigated(
      createResourceAndNavigate(
        client,
        ResourceKind.Canvas,
        undefined,
        "custom/folder",
      ),
      "/custom/folder/canvas_1.yaml",
    );
    expect(getResource.mock.calls[0][1]).toEqual({
      name: { kind: ResourceKind.Canvas, name: "canvas_1" },
    });
  });

  it("preserves PutFile failures without navigation", async () => {
    const error = new Error("write denied");
    vi.mocked(runtimeServicePutFile).mockRejectedValueOnce(error);
    await expect(
      createResourceAndNavigate(client, ResourceKind.Canvas),
    ).rejects.toBe(error);
    expect(getResource).not.toHaveBeenCalled();
    expect(navigate).not.toHaveBeenCalled();
    expect(fireEvent).not.toHaveBeenCalled();
    expect(vi.getTimerCount()).toBe(0);
  });
});
