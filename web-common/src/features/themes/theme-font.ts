/**
 * Theme fonts
 *
 * A theme can name a font family and list font files in the project's public folder.
 * The runtime serves those files only to authenticated viewers, so CSS can't fetch them:
 * a `url()` in `@font-face` has no way to send the runtime token. Instead the files are
 * fetched with the token, like any other asset, and registered through the `FontFace` API.
 */

import { protoBase64 } from "@bufbuild/protobuf";
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

/** Identifies the project a theme belongs to: the runtime that serves it and its instance. */
export interface ThemeFontScope {
  host: string;
  instanceId: string;
}

/** A theme's font, resolved for the project the theme belongs to. */
export interface ResolvedThemeFont {
  /**
   * The CSS `font-family` value for the dashboard's text and charts.
   * Vega hands it to the canvas and SVG renderers as is.
   */
  cssFamily: string;
  /** The family the font files are registered under; undefined when there are no files to load. */
  faceFamily?: string;
  files: V1ThemeFontFile[];
}

// The families that theme font files are registered under, by project and font definition.
//
// `document.fonts` is shared by the whole page. Under the theme's own family name, the faces
// of two projects that both call their font "Brand Sans" would mix (a weight one of them lacks
// would be drawn with the other's file), and files named after Rill's own font would restyle
// the app around the dashboard. So every project and set of files gets a family of its own.
const faceFamilies = new Map<string, string>();

/**
 * Resolves a theme's font for the project the theme belongs to.
 * Returns undefined when the theme keeps Rill's font.
 * Without a scope the files can't be loaded, so only the family name is used.
 */
export function resolveThemeFont(
  font: V1ThemeFont | undefined,
  scope: ThemeFontScope | undefined,
): ResolvedThemeFont | undefined {
  const family = font?.family?.trim();
  if (!family || !FONT_FAMILY_PATTERN.test(family)) return undefined;

  const isGeneric = GENERIC_FONT_FAMILIES.has(family.toLowerCase());
  const name = isGeneric ? family.toLowerCase() : `"${family}"`;
  const files = font?.files ?? [];
  if (!scope || isGeneric || files.length === 0) {
    return { cssFamily: `${name}, ${FALLBACK_FONT_FAMILIES}`, files: [] };
  }

  const key = JSON.stringify([
    scope.host,
    scope.instanceId,
    family,
    files.map((file) => [
      file.path,
      file.weightMin,
      file.weightMax,
      !!file.italic,
    ]),
  ]);
  let faceFamily = faceFamilies.get(key);
  if (!faceFamily) {
    faceFamily = `rill-theme-font-${faceFamilies.size + 1}`;
    faceFamilies.set(key, faceFamily);
  }

  // The theme's own family name comes next, for a viewer who has the font installed while
  // the files download.
  return {
    cssFamily: `"${faceFamily}", ${name}, ${FALLBACK_FONT_FAMILIES}`,
    faceFamily,
    files,
  };
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
  faceFamily: string;
  /** Resolves when the file has been registered with the document, or has failed. */
  settled: Promise<void>;
  /** The face as an `@font-face` rule with the file inlined (see themeFontEmbedCSS). */
  embedCSS?: string;
}

// One entry per face: its family, weight range and style. A face stays registered after its
// dashboard is closed, so opening the next dashboard of the project does not download it again.
const registeredFontFiles = new Map<string, RegisteredFontFile>();

/**
 * Loads a theme font's files and registers them with the document.
 * Resolves once every file has loaded or failed. A file that fails is reported on the
 * console and leaves the fallback font in place for its weight and style.
 */
export async function loadThemeFont(
  client: RuntimeClient,
  font: ResolvedThemeFont,
): Promise<void> {
  const { faceFamily } = font;
  if (!faceFamily) return;

  await Promise.all(
    font.files.map((file) => registerFontFile(client, faceFamily, file)),
  );
}

/**
 * Returns `@font-face` rules, with the font files inlined as data URIs, for the theme font
 * that a node inside a theme boundary is set in.
 *
 * PNG and PDF exports draw a copy of the dashboard as an SVG image, which can use only the
 * fonts declared inside it. Fonts from the page's stylesheets are collected for that, but a
 * theme font is registered through the `FontFace` API and appears in no stylesheet.
 */
export function themeFontEmbedCSS(node: Element): string {
  const cssFamily = getComputedStyle(node).getPropertyValue(
    "--theme-font-family",
  );
  return [...registeredFontFiles.values()]
    .filter((file) => cssFamily.includes(`"${file.faceFamily}"`))
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
  faceFamily: string,
  file: V1ThemeFontFile,
): Promise<void> {
  const descriptors = fontFaceDescriptors(file);
  const key = `${faceFamily}|${descriptors.weight}|${descriptors.style}`;

  const existing = registeredFontFiles.get(key);
  if (existing) return existing.settled;

  const entry: RegisteredFontFile = {
    faceFamily,
    settled: Promise.resolve(),
  };
  entry.settled = fetchFontFile(client, file.path ?? "")
    .then(async (data) => {
      const face = new FontFace(faceFamily, data, {
        ...descriptors,
        display: "swap",
      });
      await face.load();
      document.fonts.add(face);
      entry.embedCSS =
        `@font-face{font-family:"${faceFamily}";font-weight:${descriptors.weight};` +
        `font-style:${descriptors.style};src:url(data:font/${fontFormat(file.path)};base64,${protoBase64.enc(new Uint8Array(data))})}`;
    })
    .catch((error: unknown) => {
      console.warn(`Failed to load theme font file "${file.path}":`, error);
      // Forget the failure, so that the next theme that names this file tries again.
      registeredFontFiles.delete(key);
    });
  registeredFontFiles.set(key, entry);

  return entry.settled;
}

function fontFormat(path: string | undefined): string {
  return path?.split(".").pop()?.toLowerCase() ?? "woff2";
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
