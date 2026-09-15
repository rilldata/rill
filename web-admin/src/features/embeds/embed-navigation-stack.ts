/**
 * SvelteKit writes a per-entry index into `history.state` on every navigation it makes
 * (`sveltekit:history`, see `@sveltejs/kit/src/runtime/client/client.js`).
 * The index increments on a push, carries over unchanged on a replace,
 * and is restored along with the entry when the tab's history is traversed.
 * It is written before the page store updates, so it is readable from a page subscription.
 */
export const SVELTEKIT_HISTORY_INDEX = "sveltekit:history";

/** SvelteKit's history index for the entry the iframe is currently on, when it has one. */
export function readSvelteKitHistoryIndex(): number | undefined {
  const index = (history.state as Record<string, unknown> | null)?.[
    SVELTEKIT_HISTORY_INDEX
  ];
  return typeof index === "number" ? index : undefined;
}

/**
 * How many entries to keep. Every dashboard state change the user makes pushes an entry, the same as
 * it pushes one onto the tab's history, so an embed left open all day would otherwise grow without
 * bound. Browsers cap their own session history at a comparable size.
 */
const MAX_ENTRIES = 100;

type EmbedNavigationEntry = {
  /** Url of the entry, as a path and search string that can be handed to `goto`. */
  url: string;
  /** SvelteKit's history index for the tab entry currently holding this url, when known. */
  historyIndex: number | undefined;
};

/**
 * The embed's own navigation history, so that `navigateBack` and `navigateForward` stay inside the iframe.
 *
 * `window.history.back()` and `forward()` cannot be used for this:
 * they traverse the tab's joint session history, not the iframe's,
 * which interleaves the host page's entries with the iframe's in chronological order.
 * A `back()` from the iframe therefore reverts whichever entry happens to be one step back,
 * which is the host's entry whenever the host navigated more recently than the embed,
 * and unloads the host page entirely when the embed is still on the dashboard it loaded with.
 * This is by design: `History.back()` traverses the *traversable* navigable (the tab),
 * and it reproduces in Chromium, WebKit and Firefox alike.
 *
 * So rather than traversing, the embed tracks the urls it has visited and moves between them with `goto`,
 * which can only ever navigate the iframe.
 * The trade-off is that a back or forward pushes a tab entry of its own instead of reusing one,
 * so the browser's back button undoes the embed's last move rather than mirroring `navigateBack`.
 */
export class EmbedNavigationStack {
  private readonly entries: EmbedNavigationEntry[] = [];
  private cursor = -1;
  /** The position `take` is navigating to, until the resulting url change is recorded. */
  private pendingPosition: number | null = null;

  /**
   * Records the url the embed is now at.
   * Called for every url change regardless of who caused it: an embed API call,
   * a user interaction inside the dashboard, or a traversal of the tab's history.
   */
  public record(url: string, historyIndex?: number) {
    if (this.recordPendingMove(url, historyIndex)) return;

    const current = this.entries[this.cursor];

    // Already at this url, e.g. a page store update that did not change the url,
    // or a replace that resolved to where it started.
    // Nothing to record, but the entry may have been replaced, so keep its index current.
    if (current?.url === url) {
      current.historyIndex = historyIndex;
      return;
    }

    if (historyIndex !== undefined) {
      // The index did not change, so this url replaced the current entry rather than adding one,
      // e.g. an explore canonicalizing its url on load. It is not a separate step to go back to.
      if (current?.historyIndex === historyIndex) {
        current.url = url;
        return;
      }

      // The index belongs to an entry already visited, so the tab was traversed back onto it,
      // e.g. the user pressed the browser's back button. Follow it with the cursor.
      const visited = this.entries.findIndex(
        (entry) => entry.historyIndex === historyIndex,
      );
      if (visited !== -1) {
        this.entries[visited].url = url;
        this.cursor = visited;
        return;
      }
    }

    // A navigation to a url that is not in the stack, so anything ahead of the cursor
    // is no longer reachable, exactly as a push truncates the tab's forward entries.
    this.entries.length = this.cursor + 1;
    this.entries.push({ url, historyIndex });

    // Drop the oldest entries once the cap is reached, so the furthest back the embed can go is the
    // last MAX_ENTRIES urls rather than everything it has ever visited.
    const overflow = Math.max(this.entries.length - MAX_ENTRIES, 0);
    this.entries.splice(0, overflow);
    this.cursor = this.entries.length - 1;
  }

  /**
   * The url `delta` steps from the current position,
   * or undefined when that would leave the embed's own history,
   * as a back on the dashboard the embed loaded with would.
   * Marks the move as in flight so that the url change it causes moves the cursor
   * instead of being recorded as a new navigation.
   */
  public take(delta: -1 | 1): string | undefined {
    const position = this.cursor + delta;
    const entry = this.entries[position];
    if (!entry) return undefined;

    this.pendingPosition = position;
    return entry.url;
  }

  /**
   * Handles the url change caused by `take`: the entry moved to is already in the stack,
   * so the cursor moves onto it rather than the stack growing.
   * `goto` navigates by pushing a tab entry, so the entry adopts that entry's history index.
   * Returns whether the url change was the expected move.
   */
  private recordPendingMove(url: string, historyIndex: number | undefined) {
    if (this.pendingPosition === null) return false;

    const position = this.pendingPosition;
    this.pendingPosition = null;

    const target = this.entries[position];
    // The navigation landed somewhere else, e.g. it was redirected, or another navigation
    // overtook it. Let the caller record it as a navigation in its own right.
    if (target?.url !== url) return false;

    target.historyIndex = historyIndex;
    this.cursor = position;
    return true;
  }
}
