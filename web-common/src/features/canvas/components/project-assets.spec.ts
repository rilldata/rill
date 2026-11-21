import {
  getProjectAssetPath,
  ProjectAssetLoader,
} from "@rilldata/web-common/features/canvas/components/project-assets";
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

  function newLoader(jwt?: string) {
    const client = new RuntimeClient({
      host: "http://runtime.test",
      instanceId: "inst",
      jwt,
      authContext: "embed",
    });
    return new ProjectAssetLoader(client);
  }

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

  it("Should release its blob URLs when destroyed", async () => {
    const loader = newLoader();
    const url = await loader.resolve("public/logo.png");

    loader.destroy();
    await Promise.resolve();

    expect(URL.revokeObjectURL).toHaveBeenCalledWith(url);
  });
});
