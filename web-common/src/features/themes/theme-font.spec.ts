import type { V1ThemeFont } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  loadThemeFont,
  themeFontEmbedCSS,
  themeFontFamily,
} from "./theme-font";

describe("themeFontFamily", () => {
  it("quotes the family and falls back to Rill's font", () => {
    expect(themeFontFamily({ family: "Brand Sans" })).toBe(
      `"Brand Sans", "Inter", sans-serif`,
    );
  });

  it("leaves a generic family unquoted, since a quoted keyword names a font", () => {
    expect(themeFontFamily({ family: "System-UI" })).toBe(
      `system-ui, "Inter", sans-serif`,
    );
  });

  it("rejects a family that would add CSS of its own", () => {
    expect(
      themeFontFamily({ family: `Brand"; } body { display: none` }),
    ).toBeUndefined();
    expect(themeFontFamily({ family: "Brand, serif" })).toBeUndefined();
    expect(themeFontFamily(undefined)).toBeUndefined();
  });
});

describe("loadThemeFont", () => {
  // jsdom has neither the FontFace API nor document.fonts.
  class FakeFontFace {
    constructor(
      public family: string,
      public source: ArrayBuffer,
      public descriptors: FontFaceDescriptors,
    ) {}
    load() {
      return this.source.byteLength > 0
        ? Promise.resolve(this)
        : Promise.reject(new Error("invalid font data"));
    }
  }

  let registered: Set<FakeFontFace>;
  let fetchMock: ReturnType<typeof vi.fn>;
  let files: Record<string, ArrayBuffer | number>;

  const client = {
    host: "https://runtime.example.com",
    instanceId: "inst",
    getJwt: () => "token",
  } as unknown as RuntimeClient;

  // Each test names its own family: registered faces outlive a dashboard by design.
  function font(family: string, ...paths: string[]): V1ThemeFont {
    return {
      family,
      files: paths.map((path) => ({ path, weightMin: 400, weightMax: 400 })),
    };
  }

  beforeEach(() => {
    registered = new Set();
    files = {};
    fetchMock = vi.fn((url: string) => {
      const file = files[url];
      return Promise.resolve(
        typeof file === "number" || file === undefined
          ? { ok: false, status: file ?? 404 }
          : { ok: true, arrayBuffer: () => Promise.resolve(file) },
      );
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal("FontFace", FakeFontFace);
    Object.defineProperty(document, "fonts", {
      configurable: true,
      value: {
        add: (face: FakeFontFace) => registered.add(face),
        delete: (face: FakeFontFace) => registered.delete(face),
      },
    });
    vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  const assetURL = (path: string) =>
    `https://runtime.example.com/v1/instances/inst/assets/${path}`;

  it("fetches each file with the runtime token and registers it under the family", async () => {
    files[assetURL("public/Brand%20Sans.woff2")] = new ArrayBuffer(8);
    files[assetURL("public/BrandSans-Variable.woff2")] = new ArrayBuffer(8);

    await loadThemeFont(client, {
      family: "Brand Sans",
      files: [
        { path: "public/Brand Sans.woff2", weightMin: 700, weightMax: 700 },
        {
          path: "public/BrandSans-Variable.woff2",
          weightMin: 100,
          weightMax: 900,
          italic: true,
        },
      ],
    });

    expect(fetchMock).toHaveBeenCalledWith(
      assetURL("public/Brand%20Sans.woff2"),
      { headers: { Authorization: "Bearer token" } },
    );
    expect(
      [...registered].map(({ family, descriptors }) => ({
        family,
        weight: descriptors.weight,
        style: descriptors.style,
      })),
    ).toEqual([
      { family: "Brand Sans", weight: "700", style: "normal" },
      { family: "Brand Sans", weight: "100 900", style: "italic" },
    ]);
  });

  it("downloads a file once, however many dashboards and panels use the theme", async () => {
    files[assetURL("public/Once.woff2")] = new ArrayBuffer(8);

    await Promise.all([
      loadThemeFont(client, font("Once", "public/Once.woff2")),
      loadThemeFont(client, font("Once", "public/Once.woff2")),
    ]);
    await loadThemeFont(client, font("Once", "public/Once.woff2"));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(registered.size).toBe(1);
  });

  it("replaces a face when the theme points it at another file", async () => {
    files[assetURL("public/Old.woff2")] = new ArrayBuffer(8);
    files[assetURL("public/New.woff2")] = new ArrayBuffer(16);

    await loadThemeFont(client, font("Swap", "public/Old.woff2"));
    await loadThemeFont(client, font("Swap", "public/New.woff2"));

    expect([...registered].map((face) => face.source.byteLength)).toEqual([16]);
  });

  it("keeps the fallback font when a file is missing or is not a font, and retries later", async () => {
    files[assetURL("public/Missing.woff2")] = 403;
    files[assetURL("public/Broken.woff2")] = new ArrayBuffer(0);
    const broken = font("Broken", "public/Missing.woff2");
    broken.files?.push({ path: "public/Broken.woff2", weightMin: 700 });

    await loadThemeFont(client, broken);

    expect(registered.size).toBe(0);
    expect(console.warn).toHaveBeenCalledTimes(2);

    files[assetURL("public/Missing.woff2")] = new ArrayBuffer(8);
    await loadThemeFont(client, font("Broken", "public/Missing.woff2"));

    expect(registered.size).toBe(1);
  });

  it("offers the loaded files to exports as inlined @font-face rules", async () => {
    // PNG and PDF exports draw the dashboard as an SVG image, which sees no font of the page.
    files[assetURL("public/Export.woff2")] = new TextEncoder().encode(
      "font",
    ).buffer;

    await loadThemeFont(client, font("Export Sans", "public/Export.woff2"));

    expect(themeFontEmbedCSS()).toContain(
      `@font-face{font-family:"Export Sans";font-weight:400;font-style:normal;src:url(data:font/woff2;base64,${btoa("font")})}`,
    );
  });

  it("does not let a theme replace the fonts of Rill's own interface", async () => {
    files[assetURL("public/Inter.woff2")] = new ArrayBuffer(8);

    await loadThemeFont(client, font("Inter", "public/Inter.woff2"));

    expect(fetchMock).not.toHaveBeenCalled();
    expect(registered.size).toBe(0);
  });
});
