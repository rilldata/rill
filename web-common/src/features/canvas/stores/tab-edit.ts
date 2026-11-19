import { isMap, isSeq, type Document } from "yaml";
import { isBlockRow, renameCopiedItems, unwrapRows } from "./block-edit";

// Pure YAML-document transforms for the tabs of a tab group in the visual editor.
// Operations on whole blocks (tab groups and containers alike: add, move, duplicate,
// rename, delete, wrap a row) live in block-edit.ts.
//
// The YAML shape of a tab group is a top-level rows entry with `tabs` (and an
// optional `name`), where each tab has a `label` (display), an optional `name`
// (URL key, defaulted from the label), and its own `rows`:
//
//   rows:
//     - name: <group>          # optional
//       tabs:
//         - name: overview     # optional URL key
//           label: Overview    # display label
//           rows: [ ... ]
//
// This differs from the proto JSON shape (row.tabGroup.tabs); see canvasRowYAML
// in runtime/parser/parse_canvas.go.

/** True if the top-level rows entry at the given index is a tab group. */
export function isTabGroupRow(doc: Document, blockIndex: number): boolean {
  return isBlockRow(doc, blockIndex, "tab-group");
}

/** Number of tabs in the tab group at the given top-level index. */
export function tabCount(doc: Document, blockIndex: number): number {
  const tabs = doc.getIn(["rows", blockIndex, "tabs"]);
  return isSeq(tabs) ? tabs.items.length : 0;
}

/** True if the tab at [blockIndex, tabIndex] has at least one row of content. */
export function tabHasContent(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
): boolean {
  const rows = doc.getIn(["rows", blockIndex, "tabs", tabIndex, "rows"]);
  return isSeq(rows) && rows.items.length > 0;
}

/**
 * Append a new empty tab to the tab group at the given top-level index.
 * Returns the index of the new tab, or -1 if the entry is not a tab group.
 */
export function addTab(doc: Document, blockIndex: number): number {
  if (!isTabGroupRow(doc, blockIndex)) return -1;
  const label = `Tab ${tabCount(doc, blockIndex) + 1}`;
  doc.addIn(["rows", blockIndex, "tabs"], doc.createNode({ label, rows: [] }));
  return tabCount(doc, blockIndex) - 1;
}

/** Set the display label of the tab at [blockIndex, tabIndex]. */
export function renameTab(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
  label: string,
): void {
  if (!isTabGroupRow(doc, blockIndex)) return;
  doc.setIn(["rows", blockIndex, "tabs", tabIndex, "label"], label);
}

/**
 * Set (or clear) the URL `name` of the tab at [blockIndex, tabIndex]. Clearing it lets the
 * parser re-derive the key from the label.
 */
export function setTabName(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
  name: string,
): void {
  if (!isTabGroupRow(doc, blockIndex)) return;
  const trimmed = name.trim();
  if (trimmed) {
    doc.setIn(["rows", blockIndex, "tabs", tabIndex, "name"], trimmed);
  } else {
    doc.deleteIn(["rows", blockIndex, "tabs", tabIndex, "name"]);
  }
}

/** Move the tab at tabIndex one position in the given direction (-1 left, 1 right). */
export function moveTab(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
  direction: -1 | 1,
): void {
  reorderTab(doc, blockIndex, tabIndex, tabIndex + direction);
}

/** Move the tab at `from` to position `to` within its group (drag-to-reorder). */
export function reorderTab(
  doc: Document,
  blockIndex: number,
  from: number,
  to: number,
): void {
  if (!isTabGroupRow(doc, blockIndex)) return;
  const tabs = doc.getIn(["rows", blockIndex, "tabs"]);
  if (!isSeq(tabs)) return;
  if (from === to || from < 0 || from >= tabs.items.length) return;
  if (to < 0 || to >= tabs.items.length) return;
  const [moved] = tabs.items.splice(from, 1);
  tabs.items.splice(to, 0, moved);
}

/**
 * Duplicate the tab at [blockIndex, tabIndex], inserting the copy immediately after it.
 * The copy's label gets a " (copy)" suffix and its `name` is dropped so the parser derives a
 * fresh URL key; inline component definitions are re-derived to fresh names too (they are
 * position-keyed). Returns the index of the new tab, or -1 if the entry is not a tab group.
 */
export function duplicateTab(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
): number {
  if (!isTabGroupRow(doc, blockIndex)) return -1;
  const tabs = doc.getIn(["rows", blockIndex, "tabs"]);
  if (!isSeq(tabs)) return -1;
  const original = tabs.items[tabIndex];
  if (!isMap(original)) return -1;

  // Drop the name so the copy gets a fresh URL key; carry everything else (rows, etc.).
  // Named items in the copy get names of their own (see renameCopiedItems).
  const { name, label, ...rest } = renameCopiedItems(
    doc,
    original.toJSON(),
  ) as {
    name?: string;
    label?: string;
    [key: string]: unknown;
  };
  void name;
  const clone = doc.createNode({
    label: `${label ?? "Tab"} (copy)`,
    ...rest,
  });
  tabs.items.splice(tabIndex + 1, 0, clone);
  return tabIndex + 1;
}

/**
 * Delete the tab at [blockIndex, tabIndex].
 *
 * If it is the group's last remaining tab, the whole group is removed and that
 * tab's rows are unwrapped back into free rows at the group's position, so no
 * layout is lost (see unwrapRows). Returns the action taken.
 */
export function deleteTab(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
): "removed-tab" | "unwrapped-group" | "noop" {
  if (!isTabGroupRow(doc, blockIndex)) return "noop";

  if (tabCount(doc, blockIndex) > 1) {
    doc.deleteIn(["rows", blockIndex, "tabs", tabIndex]);
    return "removed-tab";
  }

  // Last tab (only index 0 remains): unwrap its rows into free rows at this position.
  const unwrapped = unwrapRows(doc, blockIndex, [
    "rows",
    blockIndex,
    "tabs",
    0,
    "rows",
  ]);
  return unwrapped < 0 ? "noop" : "unwrapped-group";
}
