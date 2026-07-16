import { writable, type Readable } from "svelte/store";
import type { TocEntry } from "./toc";

export type ScrollSpy = {
  /** The id of the section currently in view. Exactly one entry is active at a time. */
  activeId: Readable<string | null>;
  /** Re-bind the observer to a new set of entries (e.g. after a tab switch re-derives the TOC). */
  setEntries: (entries: TocEntry[]) => void;
  /** Force a specific id active immediately (e.g. deep-linking on mount). */
  setActive: (id: string | null) => void;
  /**
   * Pin the active id and ignore observer updates until `unlock()`. Used during click-to-scroll so
   * the highlight doesn't flicker through every section the smooth scroll passes on its way down.
   */
  lock: (id: string) => void;
  /** Resume observer-driven tracking and settle on the section now in view. */
  unlock: () => void;
  destroy: () => void;
};

/**
 * Track which section is in view using a single IntersectionObserver rooted at the scroll
 * container (no per-frame scroll math, per the spec).
 *
 * The `rootMargin` shrinks the observed viewport to a band near the top of the container, so a
 * heading counts as "in view" once it reaches that band. The active entry is the topmost heading
 * currently in the band; when the band is empty (scrolled between two headings), the last heading
 * above the band stays active, so there is always exactly one active item.
 */
export function createScrollSpy(root: HTMLElement): ScrollSpy {
  const activeId = writable<string | null>(null);

  let entries: TocEntry[] = [];
  let observer: IntersectionObserver | undefined;
  // While locked (during a click-to-scroll), observer updates are ignored so the active item stays
  // pinned to the target instead of walking through every section the scroll passes.
  let locked = false;
  // Tracks per-id intersection state so we can pick the topmost visible heading on each change.
  const intersecting = new Map<string, boolean>();

  function recompute() {
    if (locked) return;

    // Prefer the topmost heading currently in the activation band.
    const visible = entries.filter((e) => intersecting.get(e.id));
    if (visible.length > 0) {
      visible.sort(
        (a, b) =>
          a.el.getBoundingClientRect().top - b.el.getBoundingClientRect().top,
      );
      activeId.set(visible[0].id);
      return;
    }

    // Band is empty: keep the last heading whose top is above the band (the section we're within).
    const rootTop = root.getBoundingClientRect().top;
    let candidate: string | null = entries[0]?.id ?? null;
    for (const e of entries) {
      if (e.el.getBoundingClientRect().top - rootTop <= 1) candidate = e.id;
      else break;
    }
    activeId.set(candidate);
  }

  function setEntries(next: TocEntry[]) {
    entries = next;
    intersecting.clear();
    observer?.disconnect();

    if (typeof IntersectionObserver === "undefined" || next.length === 0) {
      activeId.set(next[0]?.id ?? null);
      return;
    }

    observer = new IntersectionObserver(
      (records) => {
        for (const record of records) {
          const id = (record.target as HTMLElement).id;
          intersecting.set(id, record.isIntersecting);
        }
        recompute();
      },
      // Activation band: the top ~30% of the container.
      { root, rootMargin: "0px 0px -70% 0px", threshold: 0 },
    );

    for (const e of next) observer.observe(e.el);
    recompute();
  }

  return {
    activeId,
    setEntries,
    setActive: (id) => activeId.set(id),
    lock: (id) => {
      locked = true;
      activeId.set(id);
    },
    unlock: () => {
      locked = false;
      recompute();
    },
    destroy: () => observer?.disconnect(),
  };
}
