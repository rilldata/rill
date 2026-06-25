import type {
  V1CanvasRow,
  V1CanvasTab,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import type { CanvasEntity } from "./canvas-entity";
import { Grid } from "./grid";

/**
 * A single tab within a tab group. Owns its own Grid (the tab's rows) and the
 * YAML path prefix at which those rows live, so the existing component-path and
 * transaction machinery can target a tab's rows the same way it targets top-level rows.
 */
export class Tab {
  /** Stable identifier, derived from the label; used for URL state. */
  name: string;
  /** User-facing label. */
  displayName: string;
  /** The tab's rows. */
  grid: Grid;
  /** YAML path prefix for this tab's rows, e.g. ["rows", 2, "tabs", 0, "rows"]. */
  yamlPathPrefix: (string | number)[];

  constructor(
    canvas: CanvasEntity,
    name: string,
    displayName: string,
    yamlPathPrefix: (string | number)[],
  ) {
    this.name = name;
    this.displayName = displayName;
    this.yamlPathPrefix = yamlPathPrefix;
    this.grid = new Grid(canvas);
  }

  updateFromTab(tab: V1CanvasTab, yamlPathPrefix: (string | number)[]) {
    this.name = tab.name ?? this.name;
    this.displayName = tab.displayName ?? this.displayName;
    this.yamlPathPrefix = yamlPathPrefix;
    this.grid.updateFromCanvasRows(tab.rows ?? []);
  }
}

/**
 * A tab group: a top-level layout block that renders a strip of tabs, only one of
 * which is active (and mounted) at a time.
 */
export class TabGroup {
  /** Stable identifier; used for URL state. */
  name: string;
  /** Index of the active (mounted) tab. Editor-local while editing; URL-driven in view mode. */
  activeTabIndex = writable<number>(0);
  /** The tabs in this group. */
  tabs = writable<Tab[]>([]);
  /* A tab index to activate as soon as it exists in the spec (used after add/duplicate). */
  private pendingActiveTabIndex: number | null = null;
  /* A tab name to activate once it appears in the spec (used after reorder, where the
     destination index isn't known until the spec reflects the move). */
  private pendingActiveTabName: string | null = null;

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
    // Match existing Tab instances by their stable name, not by index. On reorder a tab keeps
    // its own grid and component instances and simply moves to a new index; matching by index
    // would instead repurpose the object at each slot, leaving stale content (a tab showing
    // its neighbour's widgets) until a full reload.
    const byName = new Map(get(this.tabs).map((t) => [t.name, t]));

    const next = tabs.map((tab, tabIndex) => {
      // NOTE: this is the YAML path (row.tabs[t].rows), which differs from the
      // proto JSON shape (row.tabGroup.tabs[t].rows). pathInYAML edits the YAML document.
      const prefix = ["rows", blockIndex, "tabs", tabIndex, "rows"];
      const tabName = tab.name ?? `tab-${tabIndex}`;
      const t =
        byName.get(tabName) ??
        new Tab(
          this.canvas,
          tabName,
          tab.displayName ?? `Tab ${tabIndex + 1}`,
          prefix,
        );
      // Always sync from the spec — a newly-created tab must populate its grid too,
      // otherwise its rows render empty until the next reprocess.
      t.updateFromTab(tab, prefix);
      return t;
    });

    this.tabs.set(next);

    // Activate a pending tab by name (after a reorder) or index (after add/duplicate), once it
    // exists in the spec.
    if (this.pendingActiveTabName !== null) {
      const index = next.findIndex((t) => t.name === this.pendingActiveTabName);
      if (index !== -1) {
        this.activeTabIndex.set(index);
        this.pendingActiveTabName = null;
      }
    } else if (
      this.pendingActiveTabIndex !== null &&
      this.pendingActiveTabIndex < next.length
    ) {
      this.activeTabIndex.set(this.pendingActiveTabIndex);
      this.pendingActiveTabIndex = null;
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
    this.pendingActiveTabName = null;
  }

  /**
   * Request that the tab with the given name become active once the spec reflects it. Used
   * after a reorder, where the moved tab keeps its name but its destination index only becomes
   * known after reconcile.
   */
  activateByNameWhenReady(tabName: string) {
    this.pendingActiveTabName = tabName;
    this.pendingActiveTabIndex = null;
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

  getActiveTab(): Tab | undefined {
    return get(this.tabs)[get(this.activeTabIndex)];
  }
}

/**
 * A top-level layout block. The canvas body is an ordered list of these: each is
 * either a plain row or a tab group, mirroring the heterogeneous `rows` array in the spec.
 */
export type LayoutBlock =
  | { kind: "row"; rowIndex: number; freeRowIndex: number }
  | { kind: "tab-group"; rowIndex: number; group: TabGroup };

/** True if the spec contains any tab groups. */
export function specHasTabGroups(rows: V1CanvasRow[] | undefined): boolean {
  return !!rows?.some((row) => !!row.tabGroup);
}
