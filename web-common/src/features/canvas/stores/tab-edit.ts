import { isMap, isSeq, type Document } from "yaml";

// Pure YAML-document transforms for authoring tab groups in the visual editor.
// These operate on the parsed YAML Document (the editor's source of truth) and
// mutate it in place; the caller persists the result via the file artifact.
//
// The YAML shape of a tab group is a top-level rows entry with `tabs` (and an
// optional `name`), where each tab has a `label` and its own `rows`:
//
//   rows:
//     - name: <group>          # optional
//       tabs:
//         - label: Overview
//           rows: [ ... ]
//
// This differs from the proto JSON shape (row.tabGroup.tabs); see canvasRowYAML
// in runtime/parser/parse_canvas.go.

const MAX_ITEMS_PER_ROW = 4;

/**
 * Move a component item between two row containers (top-level rows or a tab's rows),
 * identified by their YAML paths. Used for cross-container drags (e.g. dragging a widget
 * from the free canvas into a tab). The item is removed from the source and inserted into
 * the destination:
 *   - if `dest.col` is a number and `dest.row` points at an existing destination row with
 *     room, the item joins that row at that column (e.g. dropping to the left of a widget);
 *   - otherwise it becomes a new row inserted at `dest.row` (or appended).
 *
 * Node references for both containers are resolved up front, so removing the source row
 * never invalidates the destination even when the source sits above a destination tab group.
 * Returns true if the move was applied.
 */
export function moveItemAcrossContainers(
  doc: Document,
  source: { rowsPath: (string | number)[]; row: number; col: number },
  dest: { rowsPath: (string | number)[]; row?: number; col: number | null },
): boolean {
  const sourceSeq = doc.getIn(source.rowsPath);
  const destSeq = doc.getIn(dest.rowsPath);
  if (!isSeq(sourceSeq) || !isSeq(destSeq)) return false;

  const sourceRow = sourceSeq.items[source.row];
  if (!isMap(sourceRow)) return false;
  const sourceItems = sourceRow.get("items");
  if (!isSeq(sourceItems)) return false;
  const itemNode = sourceItems.items[source.col];
  if (itemNode === undefined) return false;

  // Decide the destination shape before mutating, so a full row falls back to a new row.
  const destRowNode =
    dest.row !== undefined ? destSeq.items[dest.row] : undefined;
  const destRowItems = isMap(destRowNode)
    ? destRowNode.get("items")
    : undefined;
  const joinExistingRow =
    dest.col !== null &&
    isSeq(destRowItems) &&
    destRowItems.items.length < MAX_ITEMS_PER_ROW;

  // Remove from the source (and drop the row if it is now empty).
  sourceItems.items.splice(source.col, 1);
  if (sourceItems.items.length === 0) {
    sourceSeq.items.splice(source.row, 1);
  }

  if (joinExistingRow && isSeq(destRowItems)) {
    const at = Math.min(dest.col as number, destRowItems.items.length);
    destRowItems.items.splice(at, 0, itemNode);
  } else {
    const newRow = doc.createNode({ items: [itemNode] });
    const at = dest.row ?? destSeq.items.length;
    destSeq.items.splice(Math.min(at, destSeq.items.length), 0, newRow);
  }

  return true;
}

/** Number of top-level entries in the rows sequence. */
function rowCount(doc: Document): number {
  const rows = doc.get("rows");
  return isSeq(rows) ? rows.items.length : 0;
}

/** True if the top-level rows entry at the given index is a tab group. */
export function isTabGroupRow(doc: Document, blockIndex: number): boolean {
  const row = doc.getIn(["rows", blockIndex]);
  return isMap(row) && row.has("tabs");
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
 * Append a new tab group (with a single empty "Tab 1") at the end of the canvas.
 * Returns the top-level index of the new group.
 */
export function addTabGroup(doc: Document): number {
  return addTabGroupAt(doc, rowCount(doc));
}

/**
 * Insert a new tab group (with a single empty "Tab 1") at the given top-level index.
 * Returns the index at which it was inserted.
 */
export function addTabGroupAt(doc: Document, index: number): number {
  const group = doc.createNode({ tabs: [{ label: "Tab 1", rows: [] }] });
  const rows = doc.get("rows");
  if (isSeq(rows)) {
    const clamped = Math.max(0, Math.min(index, rows.items.length));
    rows.items.splice(clamped, 0, group);
    return clamped;
  }
  doc.setIn(["rows"], doc.createNode([group]));
  return 0;
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

/** Rename the tab at [blockIndex, tabIndex]. */
export function renameTab(
  doc: Document,
  blockIndex: number,
  tabIndex: number,
  label: string,
): void {
  if (!isTabGroupRow(doc, blockIndex)) return;
  doc.setIn(["rows", blockIndex, "tabs", tabIndex, "label"], label);
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
 * Wrap the plain row at rowIndex into a new single-tab tab group in place. The row's
 * content becomes "Tab 1"'s only row. Returns true if the conversion happened.
 */
export function convertRowToTabGroup(doc: Document, rowIndex: number): boolean {
  const rows = doc.get("rows");
  if (!isSeq(rows)) return false;
  const row = doc.getIn(["rows", rowIndex]);
  if (!isMap(row) || row.has("tabs")) return false;

  const group = doc.createNode({
    tabs: [{ label: "Tab 1", rows: [row.toJSON()] }],
  });
  rows.items.splice(rowIndex, 1, group);
  return true;
}

/**
 * Delete the tab at [blockIndex, tabIndex].
 *
 * If it is the group's last remaining tab, the whole group is removed and that
 * tab's rows are unwrapped back into free rows at the group's position, so no
 * layout is lost. Returns the action taken.
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
  const rows = doc.get("rows");
  if (!isSeq(rows)) return "noop";

  const tabRows = doc.getIn(["rows", blockIndex, "tabs", 0, "rows"]);
  const unwrapped = isSeq(tabRows) ? tabRows.items : [];
  rows.items.splice(blockIndex, 1, ...unwrapped);

  return "unwrapped-group";
}
