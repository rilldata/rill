import DOMPurify from "dompurify";
import { marked } from "marked";

let purify: typeof DOMPurify | undefined;

/**
 * Renders the inline markdown of a dashboard banner (bold, italics, strikethrough, code and links) to sanitized HTML.
 * A banner is a single line in the app's header, so block-level markdown (headings, lists) is left as text,
 * and HTML other than inline formatting is removed.
 */
export function renderBannerMarkdown(message: string): string {
  if (!purify) {
    // A separate DOMPurify instance, so the hook doesn't apply to other users of the shared one.
    purify = DOMPurify(window);
    // Open links in a new tab: following one in place would navigate away from the dashboard, or inside an embed's iframe.
    purify.addHook("afterSanitizeAttributes", (node) => {
      if (node.nodeName !== "A") return;
      node.setAttribute("target", "_blank");
      node.setAttribute("rel", "noopener noreferrer");
    });
  }

  return purify.sanitize(marked.parseInline(message, { async: false }), {
    ALLOWED_TAGS: ["a", "b", "strong", "i", "em", "u", "s", "del", "code"],
    ALLOWED_ATTR: ["href", "title"],
  });
}
