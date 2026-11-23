import type {
  V1CanvasRow,
  V1CanvasTab,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { tabRowsPath } from "../layout-util";
import type { CanvasEntity } from "./canvas-entity";
import { RowsBlock } from "./layout-blocks";

/**
 * A tab group: a top-level layout block that renders a strip of tabs, only one of
 * which is active (and mounted) at a time.
 */
export class TabGroup {
  /** Stable identifier; used for URL state. */
  name: string;
  /** Index of the active (mounted) tab. Editor-local while editing; URL-driven in view mode.
   *  The active tab is tracked by its name across spec rebuilds (see updateFromSpec), so this
   *  index always points at the same tab after a reorder. */
  activeTabIndex = writable<number>(0);
  /** The tabs in this group, each a named rows list with a label. */
  tabs = writable<RowsBlock[]>([]);
  /* A tab index to activate as soon as it exists in the spec (used after add/duplicate, where
     the new tab's name isn't known on the client until the spec reflects it). */
  private pendingActiveTabIndex: number | null = null;

  constructor(
    private canvas: CanvasEntity,
    name: string,
  ) {
    this.name = name;
  }

  /**
   * Sync the group's tabs from the spec. The blockIndex is the top-level row index
   * at which this tab group sits, used to construct each tab's YAML path prefix.
   */
  updateFromSpec(name: string, tabs: V1CanvasTab[], blockIndex: number) {
    this.name = name;
    const current = get(this.tabs);
    // The active tab is tracked by name: capture it before the rebuild so we can re-point the
    // index at the same tab afterwards, no matter how the order changed.
    const activeName = current[get(this.activeTabIndex)]?.name;

    // Match existing Tab instances by their stable name, not by index. On reorder a tab keeps
    // its own grid and component instances and simply moves to a new index; matching by index
    // would instead repurpose the object at each slot, leaving stale content (a tab showing
    // its neighbour's widgets) until a full reload.
    const byName = new Map(current.map((t) => [t.name, t]));

    const next = tabs.map((tab, tabIndex) => {
      // NOTE: this is the YAML path (row.tabs[t].rows), which differs from the
      // proto JSON shape (row.tabGroup.tabs[t].rows). pathInYAML edits the YAML document.
      const rowsPath = tabRowsPath(blockIndex, tabIndex);
      const tabName = tab.name ?? `tab-${tabIndex}`;
      const t =
        byName.get(tabName) ??
        new RowsBlock(
          this.canvas,
          tabName,
          tab.displayName ?? `Tab ${tabIndex + 1}`,
          rowsPath,
        );
      t.name = tabName;
      t.displayName = tab.displayName ?? t.displayName;
      // Always sync from the spec — a newly-created tab must populate its grid too,
      // otherwise its rows render empty until the next reprocess.
      t.updateRows(tab.rows ?? [], rowsPath);
      return t;
    });

    this.tabs.set(next);

    if (
      this.pendingActiveTabIndex !== null &&
      this.pendingActiveTabIndex < next.length
    ) {
      // A just-added/duplicated tab: activate it by its (now-known) index.
      this.activeTabIndex.set(this.pendingActiveTabIndex);
      this.pendingActiveTabIndex = null;
    } else if (activeName !== undefined) {
      // Keep the same tab active across the rebuild. If it still exists (reorder), follow it to
      // its new index; if it was renamed (name changed) or removed, the index is left as-is and
      // clamped below, which keeps the active position stable for an in-place rename.
      const index = next.findIndex((t) => t.name === activeName);
      if (index !== -1) this.activeTabIndex.set(index);
    }

    // Clamp the active index if tabs were removed.
    const activeIndex = get(this.activeTabIndex);
    if (activeIndex >= next.length) {
      this.activeTabIndex.set(Math.max(0, next.length - 1));
    }
  }

  /**
   * Request that the tab at the given index become active once it appears in the spec.
   * Used after adding/duplicating a tab, since the spec reprocess is async.
   */
  activateWhenReady(index: number) {
    this.pendingActiveTabIndex = index;
  }

  /**
   * Optimistically update a tab's display name so the strip reflects edits as the user types,
   * before the YAML change is saved and reconciled. The committed value is reconciled later
   * via updateFromSpec.
   */
  setTabDisplayName(index: number, displayName: string) {
    const list = get(this.tabs);
    const tab = list[index];
    if (!tab || tab.displayName === displayName) return;
    tab.displayName = displayName;
    this.tabs.set([...list]);
  }

  /** Select a tab by its stable name. Returns false if no such tab exists. */
  setActiveByName(name: string): boolean {
    const index = get(this.tabs).findIndex((t) => t.name === name);
    if (index === -1) return false;
    this.activeTabIndex.set(index);
    return true;
  }

  getActiveTab(): RowsBlock | undefined {
    return get(this.tabs)[get(this.activeTabIndex)];
  }
}

/** True if the spec contains any tab groups. */
export function specHasTabGroups(rows: V1CanvasRow[] | undefined): boolean {
  return !!rows?.some((row) => !!row.tabGroup);
}
