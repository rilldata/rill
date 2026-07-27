import type { GalleryEntry } from "./example-spec";

/**
 * Bundler-resolved URLs for the vendored preview images, keyed by module path.
 * `eager` with `?url` emits only the asset URLs, not the file contents, so this
 * costs a small string map rather than inlining any image data.
 *
 * This lives apart from example-spec.ts because import.meta.glob is Vite-only,
 * and that module is also imported by the snapshot script under plain Node.
 */
const THUMBNAIL_URLS = import.meta.glob<string>("./thumbnails/*.png", {
  eager: true,
  query: "?url",
  import: "default",
});

/** The bundled URL of an entry's preview image, or undefined when it has none. */
export function thumbnailURL(entry: GalleryEntry): string | undefined {
  if (!entry.thumbnail) return undefined;
  return THUMBNAIL_URLS[`./thumbnails/${entry.thumbnail}`];
}
