import type { V1ThemeFont } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  loadThemeFont,
  resolveThemeFont,
  themeFontEmbedCSS,
} from "./theme-font";

const HOST = "https://runtime.example.com";

// Registered faces outlive a dashboard by design, so each test uses projects of its own.
let nextInstance = 0;
function newProject(jwt = "token"): RuntimeClient {
  return new RuntimeClient({
    host: HOST,
    instanceId: `inst${++nextInstance}`,
    jwt,
  });
}

function font(family: string, ...paths: string[]): V1ThemeFont {
  return {
    family,
    files: paths.map((path) => ({ path, weightMin: 400, weightMax: 400 })),
  };
}

describe("resolveThemeFont", () => {
  it("quotes the family and falls back to Rill's font", () => {
    expect(resolveThemeFont({ family: "Brand Sans" }, newProject())).toEqual({
      cssFamily: `"Brand Sans", "Inter", sans-serif`,
      files: [],
    });
  });

  it("leaves a generic family unquoted, since a quoted keyword names a font", () => {
    expect(
      resolveThemeFont({ family: "System-UI" }, newProject())?.cssFamily,
    ).toBe(`system-ui, "Inter", sans-serif`);
  });

  it("rejects a family that would add CSS of its own", () => {
    expect(
      resolveThemeFont({ family: `Brand"; } body { display: none` }, undefined),
    ).toBeUndefined();
    expect(resolveThemeFont({ family: "Brand, serif" }, undefined)).toBe(
      undefined,
    );
    expect(resolveThemeFont(undefined, undefined)).toBeUndefined();
  });

  it("gives every project and every set of files a family of its own", () => {
    const project = newProject();
    const regular = font("Brand Sans", "public/BrandSans.woff2");
    const resolved = resolveThemeFont(regular, project);

    expect(resolved?.cssFamily).toBe(
      `"${resolved?.faceFamily}", "Brand Sans", "Inter", sans-serif`,
    );
    expect(resolveThemeFont(regular, project)?.faceFamily).toBe(
      resolved?.faceFamily,
    );
    expect(resolveThemeFont(regular, newProject())?.faceFamily).not.toBe(
      resolved?.faceFamily,
    );
    expect(
      resolveThemeFont(font("Brand Sans", "public/Other.woff2"), project)
        ?.faceFamily,
    ).not.toBe(resolved?.faceFamily);
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

  const assetURL = (client: RuntimeClient, path: string) =>
    `${HOST}/v1/instances/${client.instanceId}/assets/${path}`;

  // Resolves a font for a project and loads it, as ThemeProvider does.
  async function load(client: RuntimeClient, themeFont: V1ThemeFont) {
    const resolved = resolveThemeFont(themeFont, client)!;
    await loadThemeFont(client, resolved);
    return resolved;
  }

  const facesOf = (family: string | undefined) =>
    [...registered]
      .filter((face) => face.family === family)
      .map(({ descriptors }) => `${descriptors.weight} ${descriptors.style}`);

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

  it("fetches each file with the runtime token and registers its weight and style", async () => {
    const client = newProject();
    files[assetURL(client, "public/Brand%20Sans.woff2")] = new ArrayBuffer(8);
    files[assetURL(client, "public/BrandSans-Variable.woff2")] =
      new ArrayBuffer(8);

    const resolved = await load(client, {
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
      assetURL(client, "public/Brand%20Sans.woff2"),
      { headers: { Authorization: "Bearer token" } },
    );
    expect(facesOf(resolved.faceFamily)).toEqual([
      "700 normal",
      "100 900 italic",
    ]);
  });

  it("waits for an expired token to be refreshed before fetching a file", async () => {
    // An unsigned JWT that expired a minute ago: the runtime would refuse it with a 401.
    const exp = Math.floor(Date.now() / 1000) - 60;
    const client = newProject(
      `eyJhbGciOiJub25lIn0.${btoa(JSON.stringify({ exp })).replace(/=+$/, "")}.signature`,
    );
    client.setJwtRefresher(() => Promise.resolve("fresh-token"));
    files[assetURL(client, "public/Fresh.woff2")] = new ArrayBuffer(8);

    const resolved = await load(client, font("Fresh", "public/Fresh.woff2"));

    expect(fetchMock).toHaveBeenCalledWith(
      assetURL(client, "public/Fresh.woff2"),
      { headers: { Authorization: "Bearer fresh-token" } },
    );
    expect(facesOf(resolved.faceFamily)).toEqual(["400 normal"]);
  });

  it("downloads a file once, however many dashboards and panels use the theme", async () => {
    const client = newProject();
    files[assetURL(client, "public/Once.woff2")] = new ArrayBuffer(8);

    await Promise.all([
      load(client, font("Once", "public/Once.woff2")),
      load(client, font("Once", "public/Once.woff2")),
    ]);
    await load(client, font("Once", "public/Once.woff2"));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(registered.size).toBe(1);
  });

  it("keeps the faces of two projects apart when their fonts share a name", async () => {
    // A weight that one project's font lacks must not be drawn with another project's file.
    const projectA = newProject();
    const projectB = newProject();
    files[assetURL(projectA, "public/Regular.woff2")] = new ArrayBuffer(8);
    files[assetURL(projectA, "public/Bold.woff2")] = new ArrayBuffer(8);
    files[assetURL(projectB, "public/Regular.woff2")] = new ArrayBuffer(8);
    const fontA = font("Brand Sans", "public/Regular.woff2");
    fontA.files?.push({ path: "public/Bold.woff2", weightMin: 700 });

    const a = await load(projectA, fontA);
    const b = await load(projectB, font("Brand Sans", "public/Regular.woff2"));

    expect(facesOf(a.faceFamily)).toEqual(["400 normal", "700 normal"]);
    expect(facesOf(b.faceFamily)).toEqual(["400 normal"]);
    expect(b.cssFamily).not.toContain(`"${a.faceFamily}"`);
  });

  it("keeps the fallback font when a file is missing or is not a font, and retries later", async () => {
    const client = newProject();
    files[assetURL(client, "public/Missing.woff2")] = 403;
    files[assetURL(client, "public/Broken.woff2")] = new ArrayBuffer(0);
    const broken = font("Broken", "public/Missing.woff2");
    broken.files?.push({ path: "public/Broken.woff2", weightMin: 700 });

    await load(client, broken);

    expect(registered.size).toBe(0);
    expect(console.warn).toHaveBeenCalledTimes(2);

    files[assetURL(client, "public/Missing.woff2")] = new ArrayBuffer(8);
    await load(client, broken);

    expect(registered.size).toBe(1);
  });

  it("loads files for a family named like Rill's own font without replacing it", async () => {
    const client = newProject();
    files[assetURL(client, "public/Inter.woff2")] = new ArrayBuffer(8);

    const resolved = await load(client, font("Inter", "public/Inter.woff2"));

    expect(facesOf(resolved.faceFamily)).toEqual(["400 normal"]);
    expect(facesOf("Inter")).toEqual([]);
  });

  it("offers exports the faces of the theme font a node is set in, inlined", async () => {
    // PNG and PDF exports draw the dashboard as an SVG image, which sees no font of the page.
    const client = newProject();
    files[assetURL(client, "public/Export.woff2")] = new TextEncoder().encode(
      "font",
    ).buffer;
    files[assetURL(client, "public/Other.woff2")] = new ArrayBuffer(8);
    const exported = await load(
      client,
      font("Export Sans", "public/Export.woff2"),
    );
    await load(client, font("Other Sans", "public/Other.woff2"));

    const node = document.createElement("div");
    node.style.setProperty("--theme-font-family", exported.cssFamily);
    const css = themeFontEmbedCSS(node);

    expect(css).toBe(
      `@font-face{font-family:"${exported.faceFamily}";font-weight:400;font-style:normal;src:url(data:font/woff2;base64,${btoa("font")})}`,
    );
  });
});
