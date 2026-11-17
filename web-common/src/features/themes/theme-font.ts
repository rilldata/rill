/**
 * Theme fonts
 *
 * A theme can name a font family and list font files in the project's public folder.
 * The runtime serves those files only to authenticated viewers, so CSS can't fetch them:
 * a `url()` in `@font-face` has no way to send the runtime token. Instead the files are
 * fetched with the token, like any other asset, and registered through the `FontFace` API.
 */

import type {
  V1ThemeFont,
  V1ThemeFontFile,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { getContext } from "svelte";
import { readable, type Readable } from "svelte/store";

/** Context key under which ThemeProvider exposes the font that charts should use. */
export const THEME_CHART_FONT_CONTEXT_KEY = Symbol("theme-chart-font");

/** The fonts behind a theme's font, for the time it loads and for glyphs it lacks. */
const FALLBACK_FONT_FAMILIES = `"Inter", sans-serif`;

// The parser enforces the same rule. It is repeated here because the name ends up in a
// stylesheet, where anything but a plain name could close the declaration and add its own.
const FONT_FAMILY_PATTERN = /^[\p{L}\p{N}][\p{L}\p{N} ._-]*$/u;

// CSS keywords for a font the browser picks. They only work unquoted.
const GENERIC_FONT_FAMILIES = new Set([
  "serif",
  "sans-serif",
  "monospace",
  "cursive",
  "fantasy",
  "system-ui",
  "ui-serif",
  "ui-sans-serif",
  "ui-monospace",
  "ui-rounded",
]);

// Families that Rill's own interface is set in. `document.fonts` is shared by the whole
// page, so files registered under these names would restyle the app around the dashboard.
const RESERVED_FONT_FAMILIES = new Set(["inter", "source code variable"]);

/**
 * Returns the CSS `font-family` value for a theme font, or undefined if the theme has none.
 * Charts take the same value, since Vega hands it to the canvas and SVG renderers as is.
 */
export function themeFontFamily(
  font: V1ThemeFont | undefined,
): string | undefined {
  const family = font?.family?.trim();
  if (!family || !FONT_FAMILY_PATTERN.test(family)) return undefined;

  const name = GENERIC_FONT_FAMILIES.has(family.toLowerCase())
    ? family.toLowerCase()
    : `"${family}"`;
  return `${name}, ${FALLBACK_FONT_FAMILIES}`;
}

/**
 * Returns the font that charts inside the current theme boundary should use.
 * It is undefined until the theme's font files have loaded: Vega measures labels when it
 * lays out a chart, so it must not see the family before the browser can draw with it.
 *
 * Must be called during component initialization.
 */
export function getThemeChartFont(): Readable<string | undefined> {
  return (
    getContext<Readable<string | undefined> | undefined>(
      THEME_CHART_FONT_CONTEXT_KEY,
    ) ?? readable(undefined)
  );
}

interface RegisteredFontFile {
  /** Where the bytes came from: the runtime, the instance and the file's path. */
  source: string;
  /** Resolves when the file has been registered with the document, or has failed. */
  settled: Promise<void>;
  face?: FontFace;
  /** The face as an `@font-face` rule with the file inlined (see themeFontEmbedCSS). */
  embedCSS?: string;
}

// One entry per face of a family: its weight range and style. A face stays registered after
// its dashboard is closed, so opening the next dashboard of the project does not download
// it again. A different file for the same face replaces it.
const registeredFontFiles = new Map<string, RegisteredFontFile>();

/**
 * Loads a theme font's files and registers them with the document.
 * Resolves once every file has loaded or failed. A file that fails is reported on the
 * console and leaves the fallback font in place for its weight and style.
 */
export async function loadThemeFont(
  client: RuntimeClient,
  font: V1ThemeFont,
): Promise<void> {
  const family = font.family?.trim();
  const files = font.files ?? [];
  if (!family || !themeFontFamily(font) || files.length === 0) return;

  if (RESERVED_FONT_FAMILIES.has(family.toLowerCase())) {
    console.warn(
      `Theme font files for "${family}" were not loaded: Rill's own interface uses that family name.`,
    );
    return;
  }

  await Promise.all(
    files.map((file) => registerFontFile(client, family, file)),
  );
}

/**
 * Returns `@font-face` rules for the theme fonts registered with the document, with the
 * font files inlined as data URIs.
 *
 * PNG and PDF exports draw a copy of the dashboard as an SVG image, which can use only the
 * fonts declared inside it. Fonts from the page's stylesheets are collected for that, but a
 * theme font is registered through the `FontFace` API and appears in no stylesheet.
 */
export function themeFontEmbedCSS(): string {
  return [...registeredFontFiles.values()]
    .map((file) => file.embedCSS ?? "")
    .join("");
}

/** Returns the `@font-face` descriptors of a font file. */
function fontFaceDescriptors(file: V1ThemeFontFile): {
  weight: string;
  style: string;
} {
  const min = file.weightMin || 400;
  const max = file.weightMax || min;
  return {
    weight: min === max ? `${min}` : `${min} ${max}`,
    style: file.italic ? "italic" : "normal",
  };
}

function registerFontFile(
  client: RuntimeClient,
  family: string,
  file: V1ThemeFontFile,
): Promise<void> {
  const descriptors = fontFaceDescriptors(file);
  const key = `${family}|${descriptors.weight}|${descriptors.style}`;
  const source = `${client.host}|${client.instanceId}|${file.path}`;

  const existing = registeredFontFiles.get(key);
  if (existing?.source === source) return existing.settled;

  const entry: RegisteredFontFile = { source, settled: Promise.resolve() };
  entry.settled = fetchFontFile(client, file.path ?? "")
    .then(async (data) => {
      const face = new FontFace(family, data, {
        ...descriptors,
        display: "swap",
      });
      await face.load();

      // A newer theme may have claimed this face while the file was downloading.
      if (registeredFontFiles.get(key) !== entry) return;
      if (existing?.face) document.fonts.delete(existing.face);
      document.fonts.add(face);
      entry.face = face;
      entry.embedCSS =
        `@font-face{font-family:"${family}";font-weight:${descriptors.weight};` +
        `font-style:${descriptors.style};src:url(data:font/${fontFormat(file.path)};base64,${toBase64(data)})}`;
    })
    .catch((error: unknown) => {
      console.warn(`Failed to load theme font file "${file.path}":`, error);
      if (registeredFontFiles.get(key) !== entry) return;
      // Forget the failure, so that the next theme that names this file tries again.
      if (existing) registeredFontFiles.set(key, existing);
      else registeredFontFiles.delete(key);
    });
  registeredFontFiles.set(key, entry);

  return entry.settled;
}

function fontFormat(path: string | undefined): string {
  return path?.split(".").pop()?.toLowerCase() ?? "woff2";
}

function toBase64(data: ArrayBuffer): string {
  const bytes = new Uint8Array(data);
  let binary = "";
  // In chunks: a font file has more bytes than a call can take as arguments.
  for (let i = 0; i < bytes.length; i += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
  }
  return btoa(binary);
}

async function fetchFontFile(
  client: RuntimeClient,
  path: string,
): Promise<ArrayBuffer> {
  const encodedPath = path.split("/").map(encodeURIComponent).join("/");
  const headers: Record<string, string> = {};
  const jwt = client.getJwt();
  if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

  const response = await fetch(
    `${client.host}/v1/instances/${client.instanceId}/assets/${encodedPath}`,
    { headers },
  );
  if (!response.ok) throw new Error(`HTTP status ${response.status}`);
  return response.arrayBuffer();
}
