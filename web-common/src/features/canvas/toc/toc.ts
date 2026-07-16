/**
 * Table-of-contents derivation for canvas dashboards.
 *
 * Canvas has no section-header primitive: section titles are authored inside `markdown`
 * components, which render `<h1>`–`<h3>` via `{@html}` (see components/markdown/Markdown.svelte).
 * The TOC is therefore derived from the headings actually rendered in the DOM, so it stays in
 * sync with dashboard content. Only the active tab is mounted, so scanning the live content
 * container naturally scopes the TOC to the active tab(s).
 */

/** Extra top offset applied when smooth-scrolling to a heading, so it doesn't sit flush against
 *  the top edge of the scroll container. */
const SCROLL_MARGIN_TOP_PX = 16;

/** Markdown heading levels the TOC is built from. */
const HEADING_SELECTOR = ".canvas-markdown :is(h1, h2, h3)";

export type TocEntry = {
  /** Anchor id, assigned to the heading element if it lacked one. */
  id: string;
  /** Full heading text; also the accessible name (never truncated in the DOM). */
  text: string;
  /** Nesting depth relative to the shallowest heading present: 0 for top-level, incrementing by
   *  one per heading level (so h2 and h3 indent differently). */
  depth: number;
  /** The heading element itself, used as the scroll target and the IntersectionObserver target. */
  el: HTMLElement;
};

/** Turn arbitrary heading text into a URL-hash-safe slug. */
export function slugify(text: string): string {
  return (
    text
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "") || "section"
  );
}

/**
 * Scan `root` for markdown headings in document order and build the TOC entries.
 *
 * Side effects on each heading element: assigns a deduped `id` if missing (so anchors and the
 * URL hash are stable) and sets `scroll-margin-top` (so smooth-scroll lands below the container
 * edge). Both are idempotent, so re-deriving after a content change is safe.
 */
export function deriveTocEntries(
  root: HTMLElement | null | undefined,
): TocEntry[] {
  if (!root) return [];

  const headings = Array.from(
    root.querySelectorAll<HTMLElement>(HEADING_SELECTOR),
  )
    .map((el) => ({
      el,
      text: el.textContent?.trim() ?? "",
      level: Number(el.tagName[1]) || 1, // "H2" -> 2
    }))
    .filter((h) => h.text);

  if (headings.length === 0) return [];

  // The shallowest level present becomes the top level; anything deeper nests one indent per level.
  const minLevel = Math.min(...headings.map((h) => h.level));

  const usedIds = new Set<string>();

  return headings.map(({ el, text, level }) => {
    let id = el.id || slugify(text);
    // Dedupe collisions with a numeric suffix, e.g. "overview", "overview-2".
    if (usedIds.has(id)) {
      let n = 2;
      while (usedIds.has(`${id}-${n}`)) n++;
      id = `${id}-${n}`;
    }
    usedIds.add(id);

    if (el.id !== id) el.id = id;
    el.style.scrollMarginTop = `${SCROLL_MARGIN_TOP_PX}px`;

    return { id, text, depth: level - minLevel, el };
  });
}
