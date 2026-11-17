import DOMPurify from "dompurify";
import {
  getProjectAssetPath,
  type ProjectAssetLoader,
} from "../project-assets";

// Holds the project-relative path of an image that is waiting to be loaded.
export const ASSET_SRC_ATTRIBUTE = "data-asset-src";

let purify: typeof DOMPurify | undefined;

/**
 * Sanitizes rendered markdown, and moves the `src` of every image that points at a project asset to a data attribute.
 * The browser would otherwise request it relative to the page, without the user's token; `loadAssetImages` loads it instead.
 */
export function sanitizeMarkdownHtml(html: string): string {
  if (!purify) {
    // A separate DOMPurify instance, so the hook doesn't apply to other users of the shared one.
    purify = DOMPurify(window);
    purify.addHook("afterSanitizeAttributes", (node) => {
      if (node.nodeName !== "IMG") return;
      const path = getProjectAssetPath(node.getAttribute("src") ?? "");
      if (path === null) return;
      node.removeAttribute("src");
      node.setAttribute(ASSET_SRC_ATTRIBUTE, path);
    });
  }
  return purify.sanitize(html);
}

/**
 * Svelte action that loads the images `sanitizeMarkdownHtml` set aside, as they are added to the node.
 * An image keeps its data attribute until it has loaded or failed, so it can be hidden in the meantime.
 */
export function loadAssetImages(node: HTMLElement, loader: ProjectAssetLoader) {
  const loading = new WeakSet<HTMLImageElement>();

  function load() {
    const images = node.querySelectorAll<HTMLImageElement>(
      `img[${ASSET_SRC_ATTRIBUTE}]`,
    );
    for (const img of images) {
      if (loading.has(img)) continue;
      loading.add(img);
      loader
        .resolve(img.getAttribute(ASSET_SRC_ATTRIBUTE) ?? "")
        .then((src) => (img.src = src))
        // Without a source, the browser shows the image's alt text.
        .catch(() => {})
        .finally(() => img.removeAttribute(ASSET_SRC_ATTRIBUTE));
    }
  }

  load();
  const observer = new MutationObserver(load);
  observer.observe(node, { childList: true, subtree: true });

  return { destroy: () => observer.disconnect() };
}
