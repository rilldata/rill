import {
  loadAssetImages,
  sanitizeMarkdownHtml,
} from "@rilldata/web-common/features/canvas/components/markdown/asset-images";
import { ProjectAssetLoader } from "@rilldata/web-common/features/canvas/components/project-assets";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import DOMPurify from "dompurify";
import { marked } from "marked";
import { describe, expect, it, vi } from "vitest";

function render(markdown: string): HTMLElement {
  const container = document.createElement("div");
  container.innerHTML = sanitizeMarkdownHtml(
    marked.parse(markdown, { async: false }),
  );
  return container;
}

describe("sanitizeMarkdownHtml", () => {
  it("Should set aside images that point at project assets, in markdown and in HTML", () => {
    const container = render(
      [
        "![Logo](public/img/logo.svg)",
        '<img src="/v1/instances/default/assets/public/wide.png" height="40" alt="Wide">',
      ].join("\n\n"),
    );

    const [logo, wide] = container.querySelectorAll("img");
    expect(logo.hasAttribute("src")).toBe(false);
    expect(logo.getAttribute("data-asset-src")).toBe("public/img/logo.svg");
    expect(logo.alt).toBe("Logo");
    expect(wide.hasAttribute("src")).toBe(false);
    expect(wide.getAttribute("data-asset-src")).toBe("public/wide.png");
    expect(wide.getAttribute("height")).toBe("40");
  });

  it("Should leave images with an absolute URL alone", () => {
    const container = render("![Logo](https://example.com/logo.png)");

    const img = container.querySelector("img");
    expect(img?.getAttribute("src")).toBe("https://example.com/logo.png");
    expect(img?.hasAttribute("data-asset-src")).toBe(false);
  });

  it("Should still remove unsafe HTML", () => {
    const container = render(
      '<img src="public/logo.png" onerror="alert(1)"><script>alert(1)</script>',
    );

    expect(container.querySelector("script")).toBeNull();
    expect(container.querySelector("img")?.hasAttribute("onerror")).toBe(false);
  });

  it("Should not change how the shared DOMPurify instance sanitizes", () => {
    render("![Logo](public/logo.png)");

    expect(DOMPurify.sanitize('<img src="public/logo.png">')).toBe(
      '<img src="public/logo.png">',
    );
  });
});

describe("loadAssetImages", () => {
  // Loads asset paths like ProjectAssetLoader does, without a runtime to fetch from.
  function fakeLoader() {
    const requested: string[] = [];
    const loader = {
      load: (path: string) => {
        requested.push(path);
        return path.includes("missing")
          ? Promise.reject(new Error("HTTP 404"))
          : Promise.resolve(`blob:test/${path}`);
      },
    } as unknown as ProjectAssetLoader;
    return { loader, requested };
  }

  // Lets the MutationObserver and the loader's promises run.
  const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

  it("Should load the images that are in the node, and the ones rendered later", async () => {
    const { loader, requested } = fakeLoader();
    const node = render("![Logo](public/logo.png)");
    const action = loadAssetImages(node, loader);
    await settle();

    let img = node.querySelector("img");
    expect(img?.getAttribute("src")).toBe("blob:test/public/logo.png");
    expect(img?.hasAttribute("data-asset-src")).toBe(false);

    // A re-render replaces the content, as `{@html}` does.
    node.innerHTML = render("![Icon](public/img/icon.svg)").innerHTML;
    await settle();

    img = node.querySelector("img");
    expect(img?.getAttribute("src")).toBe("blob:test/public/img/icon.svg");
    expect(requested).toEqual(["public/logo.png", "public/img/icon.svg"]);

    action.destroy();
  });

  it("Should leave an image that fails to load without a source, so its alt text shows", async () => {
    const { loader } = fakeLoader();
    const node = render("![Missing logo](public/missing.png)");
    const action = loadAssetImages(node, loader);
    await settle();

    const img = node.querySelector("img");
    expect(img?.hasAttribute("src")).toBe(false);
    expect(img?.hasAttribute("data-asset-src")).toBe(false);
    expect(img?.alt).toBe("Missing logo");

    action.destroy();
  });

  it("Should only fetch a set-aside path from the runtime, never use it as the source", async () => {
    const fetchMock = vi.fn<typeof fetch>(() =>
      Promise.resolve(new Response("", { status: 403 })),
    );
    vi.stubGlobal("fetch", fetchMock);
    const loader = new ProjectAssetLoader(
      new RuntimeClient({
        host: "http://runtime.test",
        instanceId: "inst",
        authContext: "embed",
      }),
    );
    const node = render(
      [
        // The sanitizer sets aside what follows the assets API path, which can be a URL.
        '<img src="/v1/instances/default/assets/javascript:alert(1)" alt="Script">',
        // The data attribute itself can be written in the markdown.
        '<img data-asset-src="https://tracker.example/pixel.png" alt="Tracker">',
      ].join("\n\n"),
    );
    const action = loadAssetImages(node, loader);
    await settle();

    expect(fetchMock.mock.calls.map(([url]) => url)).toEqual([
      "http://runtime.test/v1/instances/inst/assets/javascript%3Aalert(1)",
      "http://runtime.test/v1/instances/inst/assets/https%3A/tracker.example/pixel.png",
    ]);
    for (const img of node.querySelectorAll("img")) {
      expect(img.hasAttribute("src")).toBe(false);
    }

    action.destroy();
    vi.unstubAllGlobals();
  });
});
