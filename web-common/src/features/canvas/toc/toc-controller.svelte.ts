import { deriveTocEntries, type TocEntry } from "./toc";

// Activation band: a heading counts as "in view" once it reaches the top ~30% of the container.
const ACTIVATION_ROOT_MARGIN = "0px 0px -70% 0px";
// Debounce for re-deriving after DOM changes (tab switch, async markdown, edits).
const REFRESH_DEBOUNCE_MS = 150;
// Idle gap that counts as "scrolling stopped", and a hard cap in case no scroll events fire.
const SCROLL_IDLE_MS = 100;
const SCROLL_MAX_WAIT_MS = 1000;

/**
 * Owns the table-of-contents state and behaviour for a canvas dashboard.
 *
 * Canvas has no section-header primitive, so entries are derived from the `<h1>`–`<h3>` rendered
 * inside `markdown` components of the active tab (see `deriveTocEntries`) and kept in sync via a
 * MutationObserver. The section in view is tracked with a single IntersectionObserver (no per-frame
 * scroll math): the active entry is the topmost heading in the activation band, falling back to the
 * last heading above it when the band is empty, so exactly one entry is active at a time.
 */
export class CanvasTocController {
  /** Sections derived from the active tab's rendered headings. */
  entries = $state<TocEntry[]>([]);
  /** The section currently in view; exactly one at a time. */
  activeId = $state<string | null>(null);

  private observer: IntersectionObserver | undefined;
  private mutationObserver: MutationObserver | undefined;
  // Per-id intersection state, so we can pick the topmost visible heading on each change.
  private readonly intersecting = new Map<string, boolean>();
  // While locked (during a click-to-scroll), observer updates are ignored so the active item stays
  // pinned to the target instead of walking through every section the scroll passes.
  private locked = false;
  private deepLinked = false;
  private refreshTimer: ReturnType<typeof setTimeout> | undefined;
  private cancelScrollWatch: (() => void) | undefined;

  constructor(private readonly scrollContainer: HTMLElement) {
    this.refresh();

    // Re-derive on any content change: tab switch, async markdown resolution, edits, filters, or the
    // `.row-container` appearing (e.g. after required filters are satisfied). Observing the scroll
    // container rather than `.row-container` keeps this resilient to remounts.
    this.mutationObserver = new MutationObserver(() => this.scheduleRefresh());
    this.mutationObserver.observe(scrollContainer, {
      childList: true,
      subtree: true,
      characterData: true,
    });
  }

  /** Smooth-scroll to a section on click, update the URL hash, and pin the highlight to it. */
  jumpTo(event: MouseEvent, entry: TocEntry) {
    // Handle navigation ourselves so SvelteKit doesn't intercept the hash link, and so we can
    // respect prefers-reduced-motion.
    event.preventDefault();
    // Pin the highlight to the target for the duration of the scroll, so it doesn't flicker through
    // every section the smooth scroll passes; release once the scroll settles.
    this.locked = true;
    this.activeId = entry.id;
    entry.el.scrollIntoView({
      behavior: this.scrollBehavior(),
      block: "start",
    });
    history.replaceState(history.state, "", `#${entry.id}`);
    this.watchScrollEnd();
    // Release focus so the flyout collapses once the pointer leaves; otherwise `:focus-within`
    // keeps it open until the user clicks elsewhere.
    (event.currentTarget as HTMLElement).blur();
  }

  destroy() {
    clearTimeout(this.refreshTimer);
    this.cancelScrollWatch?.();
    this.mutationObserver?.disconnect();
    this.observer?.disconnect();
  }

  private scheduleRefresh() {
    clearTimeout(this.refreshTimer);
    this.refreshTimer = setTimeout(() => this.refresh(), REFRESH_DEBOUNCE_MS);
  }

  private refresh() {
    const rowContainer =
      this.scrollContainer.querySelector<HTMLElement>(".row-container");
    this.entries = deriveTocEntries(rowContainer);
    this.observe();

    // On first render, honor a deep link to a section (tab selection is handled separately via the
    // existing `tabs` URL param). Use an instant jump so the page doesn't animate on load.
    if (!this.deepLinked && this.entries.length > 0) {
      this.deepLinked = true;
      const hash = decodeURIComponent(window.location.hash.slice(1));
      const target = hash && this.entries.find((e) => e.id === hash);
      if (target) {
        target.el.scrollIntoView({ behavior: "auto", block: "start" });
        this.activeId = target.id;
      }
    }
  }

  private observe() {
    this.intersecting.clear();
    this.observer?.disconnect();

    if (
      typeof IntersectionObserver === "undefined" ||
      this.entries.length === 0
    ) {
      this.activeId = this.entries[0]?.id ?? null;
      return;
    }

    this.observer = new IntersectionObserver(
      (records) => {
        for (const record of records) {
          this.intersecting.set(
            (record.target as HTMLElement).id,
            record.isIntersecting,
          );
        }
        this.recompute();
      },
      {
        root: this.scrollContainer,
        rootMargin: ACTIVATION_ROOT_MARGIN,
        threshold: 0,
      },
    );
    for (const entry of this.entries) this.observer.observe(entry.el);
    this.recompute();
  }

  private recompute() {
    if (this.locked) return;

    // Prefer the topmost heading currently in the activation band.
    const visible = this.entries.filter((e) => this.intersecting.get(e.id));
    if (visible.length > 0) {
      visible.sort(
        (a, b) =>
          a.el.getBoundingClientRect().top - b.el.getBoundingClientRect().top,
      );
      this.activeId = visible[0].id;
      return;
    }

    // Band is empty: keep the last heading whose top is above it (the section we're within).
    const rootTop = this.scrollContainer.getBoundingClientRect().top;
    let candidate: string | null = this.entries[0]?.id ?? null;
    for (const entry of this.entries) {
      if (entry.el.getBoundingClientRect().top - rootTop <= 1)
        candidate = entry.id;
      else break;
    }
    this.activeId = candidate;
  }

  private scrollBehavior(): ScrollBehavior {
    return window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
      ? "auto"
      : "smooth";
  }

  // Unlock once the container stops scrolling (debounced), with a hard timeout in case no scroll
  // events fire (e.g. the target is already in view, or reduced-motion jumps instantly).
  private watchScrollEnd() {
    this.cancelScrollWatch?.();

    let idle: ReturnType<typeof setTimeout>;
    const finish = () => {
      this.cancelScrollWatch?.();
      this.locked = false;
      this.recompute();
    };
    const onScroll = () => {
      clearTimeout(idle);
      idle = setTimeout(finish, SCROLL_IDLE_MS);
    };
    const maxWait = setTimeout(finish, SCROLL_MAX_WAIT_MS);

    this.scrollContainer.addEventListener("scroll", onScroll, {
      passive: true,
    });
    this.cancelScrollWatch = () => {
      clearTimeout(idle);
      clearTimeout(maxWait);
      this.scrollContainer.removeEventListener("scroll", onScroll);
      this.cancelScrollWatch = undefined;
    };
  }
}
