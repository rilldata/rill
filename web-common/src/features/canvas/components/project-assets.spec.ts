import {
  getProjectAssetPath,
  ProjectAssetLoader,
} from "@rilldata/web-common/features/canvas/components/project-assets";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("getProjectAssetPath", () => {
  it.each([
    ["public/logo.png", "public/logo.png"],
    ["public/img/brand/logo.svg", "public/img/brand/logo.svg"],
    ["/public/logo.png", "public/logo.png"],
    ["./public/logo.png", "public/logo.png"],
    ["  public/logo.png ", "public/logo.png"],
    // The assets API path that works without a token in Rill Developer, for any instance.
    ["/v1/instances/default/assets/public/img/logo.svg", "public/img/logo.svg"],
    ["v1/instances/abc-123/assets/public/logo.png", "public/logo.png"],
  ])("Should treat %s as a project asset", (url, path) => {
    expect(getProjectAssetPath(url)).toBe(path);
  });

  it.each([
    "https://example.com/logo.png",
    "HTTP://example.com/logo.png",
    "//cdn.example.com/logo.png",
    "data:image/png;base64,iVBORw0KGgo=",
    "blob:http://localhost/1234",
    "",
    "/",
  ])("Should leave %s for the browser to load", (url) => {
    expect(getProjectAssetPath(url)).toBeNull();
  });
});

describe("ProjectAssetLoader", () => {
  const fetchMock = vi.fn();
  let blobCount = 0;

  beforeEach(() => {
    blobCount = 0;
    fetchMock.mockReset();
    fetchMock.mockImplementation(() =>
      Promise.resolve(new Response("image", { status: 200 })),
    );
    vi.stubGlobal("fetch", fetchMock);
    // jsdom doesn't implement blob URLs.
    URL.createObjectURL = () => `blob:test/${++blobCount}`;
    URL.revokeObjectURL = vi.fn();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  function newLoader(jwt?: string, onChange?: () => void) {
    const client = new RuntimeClient({
      host: "http://runtime.test",
      instanceId: "inst",
      jwt,
      authContext: "embed",
    });
    return new ProjectAssetLoader(client, onChange);
  }

  const fetchedUrls = () => fetchMock.mock.calls.map(([url]) => url as string);

  it("Should fetch an asset from the instance with the user's token, once", async () => {
    const loader = newLoader("token");

    const first = await loader.resolve("public/img/logo.svg");
    const second = await loader.resolve(
      "/v1/instances/default/assets/public/img/logo.svg",
    );

    expect(first).toBe("blob:test/1");
    expect(second).toBe(first);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://runtime.test/v1/instances/inst/assets/public/img/logo.svg",
      { headers: { Authorization: "Bearer token" } },
    );
  });

  it("Should not fetch a URL the browser can load", async () => {
    const loader = newLoader();

    expect(await loader.resolve("https://example.com/logo.png")).toBe(
      "https://example.com/logo.png",
    );
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("Should fail for an asset the runtime doesn't serve, and try again next time", async () => {
    const loader = newLoader();
    fetchMock.mockImplementationOnce(() =>
      Promise.resolve(new Response("", { status: 403 })),
    );

    await expect(loader.resolve("rill.yaml")).rejects.toThrow("HTTP 403");
    expect(await loader.resolve("rill.yaml")).toBe("blob:test/1");
  });

  it("Should encode each segment of the path once", async () => {
    const loader = newLoader();

    await loader.resolve("public/img/logo #2.png");
    // Marked percent-encodes the spaces in a markdown image's URL.
    await loader.resolve("public/img/logo%20#2.png");
    await loader.resolve("public/a?b=1.png");
    // A `%` that doesn't start an escape sequence is part of the name.
    await loader.resolve("public/100%.png");

    expect(fetchedUrls()).toEqual([
      "http://runtime.test/v1/instances/inst/assets/public/img/logo%20%232.png",
      "http://runtime.test/v1/instances/inst/assets/public/a%3Fb%3D1.png",
      "http://runtime.test/v1/instances/inst/assets/public/100%25.png",
    ]);
  });

  it("Should resolve `.` and `..` segments, so a request stays in the assets API", async () => {
    const loader = newLoader();

    await loader.load("public/img/../logo.png");
    await loader.load("public/../../../v1/instances/inst/files");
    await loader.load("public/%2e%2e/%2E%2E/health");

    expect(fetchedUrls()).toEqual([
      "http://runtime.test/v1/instances/inst/assets/public/logo.png",
      "http://runtime.test/v1/instances/inst/assets/v1/instances/inst/files",
      "http://runtime.test/v1/instances/inst/assets/health",
    ]);
  });

  it("Should fetch an asset again after its file changes", async () => {
    const onChange = vi.fn();
    const loader = newLoader(undefined, onChange);
    const first = await loader.resolve("public/img/logo.svg");

    // Changes to files that aren't loaded don't matter.
    eventBus.emit("project-file-changed", "/public/img/other.svg");
    expect(onChange).not.toHaveBeenCalled();

    eventBus.emit("project-file-changed", "/public/img/logo.svg");
    await Promise.resolve();
    expect(onChange).toHaveBeenCalledTimes(1);
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(first);

    const second = await loader.resolve("public/img/logo.svg");
    expect(second).not.toBe(first);
    expect(fetchMock).toHaveBeenCalledTimes(2);

    loader.destroy();
  });

  it("Should release its blob URLs when destroyed, and load nothing after", async () => {
    const onChange = vi.fn();
    const loader = newLoader(undefined, onChange);
    const url = await loader.resolve("public/logo.png");

    loader.destroy();
    await Promise.resolve();
    expect(URL.revokeObjectURL).toHaveBeenCalledWith(url);

    await expect(loader.resolve("public/logo.png")).rejects.toThrow(
      "destroyed",
    );
    eventBus.emit("project-file-changed", "/public/logo.png");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(onChange).not.toHaveBeenCalled();
  });
});
