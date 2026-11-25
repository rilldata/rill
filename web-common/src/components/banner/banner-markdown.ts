import DOMPurify from "dompurify";
import { Marked } from "marked";

// The inline formatting a banner can use in HTML, besides markdown.
const ALLOWED_TAGS = ["a", "b", "strong", "i", "em", "u", "s", "del", "code"];

// A separate instance, so the renderer doesn't apply to other markdown.
const marked = new Marked({
  renderer: {
    // Show other tags as written, like `<region>` in "Use the <region> filter", as plain-text banners did.
    html({ text }) {
      const tagName = /^<\/?([a-z][a-z0-9-]*)/i.exec(text)?.[1].toLowerCase();
      if (tagName && ALLOWED_TAGS.includes(tagName)) return text;
      return text
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;");
    },
  },
});

let purify: typeof DOMPurify | undefined;

/**
 * Renders the inline markdown of a dashboard banner (bold, italics, strikethrough, code and links) to sanitized HTML.
 * A banner is a single line in the app's header, so block-level markdown (headings, lists) is left as text,
 * and so is HTML other than inline formatting.
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
    ALLOWED_TAGS,
    ALLOWED_ATTR: ["href", "title"],
  });
}
