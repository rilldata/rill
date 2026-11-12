import {
  isAlias,
  isCollection,
  isMap,
  isScalar,
  isSeq,
  visit,
  type Document,
  type YAMLMap,
} from "yaml";

// Pure YAML-document transforms for the blocks of a canvas: the top-level rows entries
// that hold rows instead of items. They operate on the parsed YAML Document (the
// editor's source of truth) and mutate it in place; the caller persists the result via
// the file artifact.
//
//   rows:
//     - items: [ ... ]          # a plain row
//     - name: deep_dive         # a tab group: `tabs`, each with its own rows
//       if: <condition>         # optional; everything inside inherits it
//       tabs:
//         - label: Overview
//           rows: [ ... ]
//     - name: revenue_story     # a container: `rows`, shown to viewers as plain rows
//       rows: [ ... ]
//
// Blocks are top level only, and names are unique across both kinds (see parseCanvasRows
// in runtime/parser/parse_canvas.go). A block's name is its key. The YAML may leave it out,
// but the editor names every block before it edits a document (see nameBlocks), so an
// edit never changes a block's key. Tab-specific transforms live in tab-edit.ts.

export type BlockKind = "tab-group" | "container";

const BLOCK_KINDS: Record<
  BlockKind,
  {
    // The key that holds the block's content, which is also what makes a rows entry a block of this kind.
    contentKey: string;
    // The parser names an unnamed block with this prefix followed by its top-level index.
    defaultNamePrefix: string;
    // A new, empty block of this kind.
    newEntry: () => Record<string, unknown>;
    // A block of this kind whose only content is the given row.
    wrapEntry: (row: unknown) => Record<string, unknown>;
  }
> = {
  "tab-group": {
    contentKey: "tabs",
    defaultNamePrefix: "group-",
    newEntry: () => ({ tabs: [{ label: "Tab 1", rows: [] }] }),
    wrapEntry: (row) => ({ tabs: [{ label: "Tab 1", rows: [row] }] }),
  },
  container: {
    contentKey: "rows",
    defaultNamePrefix: "container-",
    newEntry: () => ({ rows: [] }),
    wrapEntry: (row) => ({ rows: [row] }),
  },
};

const BLOCK_KIND_LIST = Object.keys(BLOCK_KINDS) as BlockKind[];

const MAX_ITEMS_PER_ROW = 4;

/**
 * The kind of block the top-level rows entry at the index is, or undefined for a plain row.
 * The editor edits only blocks written without YAML aliases: moving or removing an aliased
 * block can leave an alias before its anchor, or without one, and the document can't be
 * written then. Naming follows aliases, as the parser does (see parsedBlockAt).
 */
export function blockKindAt(
  doc: Document,
  index: number,
): BlockKind | undefined {
  const entry = doc.getIn(["rows", index]);
  if (!isMap(entry)) return undefined;
  // Like the parser, a list under the content key makes the block, whatever `items` holds.
  return BLOCK_KIND_LIST.find((kind) =>
    isSeq(entry.get(BLOCK_KINDS[kind].contentKey)),
  );
}

/** True if the top-level rows entry at the index is a block, of the given kind if one is passed. */
export function isBlockRow(
  doc: Document,
  index: number,
  kind?: BlockKind,
): boolean {
  const actual = blockKindAt(doc, index);
  return actual !== undefined && (kind === undefined || actual === kind);
}

/** The name the parser gives an unnamed block at a top-level index. */
export function defaultBlockName(kind: BlockKind, index: number): string {
  return `${BLOCK_KINDS[kind].defaultNamePrefix}${index}`;
}

/** True if the name has the form of a default or new block name of the kind, such as "container-2". */
export function isGeneratedBlockName(kind: BlockKind, name: string): boolean {
  const prefix = BLOCK_KINDS[kind].defaultNamePrefix;
  return name.startsWith(prefix) && /^\d+$/.test(name.slice(prefix.length));
}

/**
 * Write the default name of every unnamed block into the document. A block's key is its
 * name: a tab group's URL key, and what the editor's selection and a block's editor state
 * (a group's active tab) follow. The default name holds the block's row index, so without
 * this an edit that shifts rows would change the keys of the blocks below. The editor calls
 * this on the documents of the edits that can shift rows, those of the canvas builder and
 * the block inspector (see CanvasBuilder and applyCanvasEdit), before the edit. Returns the
 * document.
 */
export function nameBlocks(doc: Document): Document {
  const rows = doc.get("rows");
  if (!isSeq(rows)) return doc;
  rows.items.forEach((_, index) => {
    const block = parsedBlockAt(doc, index);
    // A shared map stays unnamed: every copy of it would get the name.
    if (block && !block.shared && !blockNameOf(doc, block.map)) {
      setInMapBeforeContent(
        doc,
        block.map,
        "name",
        defaultBlockName(block.kind, index),
      );
    }
  });
  return doc;
}

/**
 * The top-level index of the block with the name, as the parser names blocks: its `name`,
 * or the default name of an unnamed block. Returns -1 if no block has it.
 */
export function blockIndexOf(doc: Document, name: string): number {
  const rows = doc.get("rows");
  if (!isSeq(rows)) return -1;
  return rows.items.findIndex(
    (_, index) => parsedBlockName(doc, index) === name,
  );
}

/**
 * Insert a new, empty block of the kind at the top-level index (a tab group starts with
 * a single empty "Tab 1"). Returns the index at which it was inserted.
 */
export function addBlockAt(
  doc: Document,
  index: number,
  kind: BlockKind,
): number {
  const entry = doc.createNode({
    name: newBlockName(doc, kind),
    ...BLOCK_KINDS[kind].newEntry(),
  });
  const rows = doc.get("rows");
  if (isSeq(rows)) {
    const clamped = Math.max(0, Math.min(index, rows.items.length));
    rows.items.splice(clamped, 0, entry);
    return clamped;
  }
  doc.setIn(["rows"], doc.createNode([entry]));
  return 0;
}

/**
 * Set the `name` of the block at the top-level index. Returns true if it was set. An empty
 * name is ignored, since every block the editor writes has a name, and so is a name another
 * block has, since the parser rejects a canvas whose tab groups and containers share one.
 */
export function renameBlock(
  doc: Document,
  index: number,
  name: string,
): boolean {
  const trimmed = name.trim();
  if (!trimmed || !isBlockRow(doc, index)) return false;
  const holder = blockIndexOf(doc, trimmed);
  if (holder !== -1 && holder !== index) return false;
  setKeyBeforeContent(doc, ["rows", index], "name", trimmed);
  return true;
}

/**
 * Set `key` of the map at `path` to `value`. A key the map doesn't have yet goes before the
 * map's content, its first key that holds a list or a map (a row's `items`, a block's `tabs`
 * or `rows`, an item's component), so short properties such as `name` and `if` read first.
 */
export function setKeyBeforeContent(
  doc: Document,
  path: (string | number)[],
  key: string,
  value: unknown,
): void {
  const map = doc.getIn(path);
  if (isMap(map)) {
    setInMapBeforeContent(doc, map, key, value);
  } else {
    doc.setIn([...path, key], value);
  }
}

function setInMapBeforeContent(
  doc: Document,
  map: YAMLMap,
  key: string,
  value: unknown,
): void {
  if (map.has(key)) {
    map.set(key, value);
    return;
  }
  const at = map.items.findIndex((pair) => isCollection(pair.value));
  map.items.splice(
    at === -1 ? map.items.length : at,
    0,
    doc.createPair(key, value),
  );
}

/**
 * Duplicate the block at the top-level index, inserting the copy immediately after it under
 * a new name (tab names are scoped to their group, so they can stay). Everything inside is
 * copied; inline component definitions are position-keyed, so the parser derives fresh names
 * for the copies. Returns the copy's name, or undefined if the entry is not a block.
 */
export function duplicateBlock(
  doc: Document,
  index: number,
): string | undefined {
  const kind = blockKindAt(doc, index);
  const rows = doc.get("rows");
  if (!kind || !isSeq(rows)) return undefined;
  const original = rows.items[index];
  if (!isMap(original)) return undefined;

  const name = newBlockName(doc, kind);
  const copy = doc.createNode(original.toJSON());
  if (!isMap(copy)) return undefined;
  setInMapBeforeContent(doc, copy, "name", name);
  rows.items.splice(index + 1, 0, copy);
  return name;
}

/**
 * Move the block at top-level index `from` to the insertion index `to` (drag-to-reorder).
 * `to` is a slot in the pre-move list, as a row drop zone reports it: 0 is above the first
 * block and rows.length is below the last. Returns the block's new index, or -1 if nothing
 * moved (not a block, out of range, or dropped in one of its own two adjacent slots).
 */
export function moveBlock(doc: Document, from: number, to: number): number {
  if (!isBlockRow(doc, from)) return -1;
  const rows = doc.get("rows");
  if (!isSeq(rows)) return -1;
  if (to < 0 || to > rows.items.length) return -1;
  if (to === from || to === from + 1) return -1;
  const [moved] = rows.items.splice(from, 1);
  const newIndex = to > from ? to - 1 : to;
  rows.items.splice(newIndex, 0, moved);
  return newIndex;
}

/**
 * Delete the block at the top-level index with everything inside it.
 * Returns true if a block was removed.
 */
export function deleteBlock(doc: Document, index: number): boolean {
  if (!isBlockRow(doc, index)) return false;
  const rows = doc.get("rows");
  if (!isSeq(rows)) return false;
  rows.items.splice(index, 1);
  return true;
}

/** True if the block at the top-level index has at least one row, in any of its tabs for a tab group. */
export function blockHasContent(doc: Document, index: number): boolean {
  const hasRows = (rows: unknown) => isSeq(rows) && rows.items.length > 0;
  switch (blockKindAt(doc, index)) {
    case "tab-group": {
      const tabs = doc.getIn(["rows", index, "tabs"]);
      return (
        isSeq(tabs) &&
        tabs.items.some((tab) => isMap(tab) && hasRows(tab.get("rows")))
      );
    }
    case "container":
      return hasRows(doc.getIn(["rows", index, "rows"]));
    default:
      return false;
  }
}

/**
 * Replace the plain row at the top-level index with a new block of the kind whose only
 * content is that row (for a tab group, the only row of "Tab 1").
 * Returns true if the row was wrapped.
 */
export function wrapRow(
  doc: Document,
  rowIndex: number,
  kind: BlockKind,
): boolean {
  const rows = doc.get("rows");
  if (!isSeq(rows)) return false;
  const row = rows.items[rowIndex];
  if (!isMap(row) || isBlockRow(doc, rowIndex)) return false;

  const block = doc.createNode({
    name: newBlockName(doc, kind),
    ...BLOCK_KINDS[kind].wrapEntry(row.toJSON()),
  });
  rows.items.splice(rowIndex, 1, block);
  return true;
}

/**
 * The top-level entry at the index as the parser reads it, through YAML aliases: the block's
 * kind and map, and whether the map is shared, as an alias or the anchor of one, so that a
 * name written in it would name every copy. Undefined for a plain row.
 */
function parsedBlockAt(
  doc: Document,
  index: number,
): { kind: BlockKind; map: YAMLMap; shared: boolean } | undefined {
  const entry: unknown = doc.getIn(["rows", index], true);
  const map = resolveAlias(doc, entry);
  if (!isMap(map)) return undefined;
  const kind = BLOCK_KIND_LIST.find((k) =>
    isSeq(resolveAlias(doc, map.get(BLOCK_KINDS[k].contentKey, true))),
  );
  if (!kind) return undefined;
  return { kind, map, shared: isAlias(entry) || map.anchor !== undefined };
}

/** The name the parser gives the block at the index, or undefined for a plain row. */
function parsedBlockName(doc: Document, index: number): string | undefined {
  const block = parsedBlockAt(doc, index);
  if (!block) return undefined;
  return blockNameOf(doc, block.map) || defaultBlockName(block.kind, index);
}

/** The `name` of a block's map, or "" when it has none and the parser gives it the default. */
function blockNameOf(doc: Document, map: YAMLMap): string {
  const node = resolveAlias(doc, map.get("name", true));
  const name: unknown = isScalar(node) ? node.value : undefined;
  // A scalar of any type is a name to the parser, which reads it as a string.
  return typeof name === "string" ||
    typeof name === "number" ||
    typeof name === "boolean"
    ? String(name)
    : "";
}

function resolveAlias(doc: Document, node: unknown): unknown {
  return isAlias(node) ? node.resolve(doc) : node;
}

/**
 * A name for a new block of the kind that no block in the document has: the kind's prefix
 * with the smallest free number from 1, such as "container-1". An unnamed block holds its
 * default name, so the editor names the blocks first (see nameBlocks), or inserting the new
 * block would shift a default name onto the new one.
 */
function newBlockName(doc: Document, kind: BlockKind): string {
  const rows = doc.get("rows");
  const taken = new Set(
    isSeq(rows)
      ? rows.items.map((_, index) => parsedBlockName(doc, index))
      : [],
  );
  let n = 1;
  while (taken.has(defaultBlockName(kind, n))) n++;
  return defaultBlockName(kind, n);
}

/**
 * Replace the block at the top-level index with the rows at rowsPath inside it (a
 * container's rows, or one tab's), so no layout is lost. Each row keeps the visibility of
 * the elements it leaves (the container's `if`, or the tab group's and the tab's), combined
 * with its own, so content hidden from some viewers stays hidden from them.
 * Returns the number of rows put in the block's place, or -1 if the entry is not a block.
 */
export function unwrapRows(
  doc: Document,
  index: number,
  rowsPath: (string | number)[],
): number {
  if (!isBlockRow(doc, index)) return -1;
  const rows = doc.get("rows");
  if (!isSeq(rows)) return -1;

  const inner = doc.getIn(rowsPath);
  const unwrapped = isSeq(inner) ? inner.items : [];

  const conditions = ancestorConditions(doc, rowsPath);
  if (combineConditions(...conditions) !== undefined) {
    for (let i = 0; i < unwrapped.length; i++) {
      const row = ownMapAt(doc, unwrapped, i);
      if (!row) continue;
      const condition = combineConditions(...conditions, row.get("if"));
      setInMapBeforeContent(doc, row, "if", condition);
    }
  }

  rows.items.splice(index, 1, ...unwrapped);
  return unwrapped.length;
}

/**
 * The `if` conditions of the elements that hold the rows at rowsPath (a container, or a tab
 * group and its tab): the prefixes of rowsPath that end at a list index. Unset ones are
 * undefined.
 */
function ancestorConditions(
  doc: Document,
  rowsPath: (string | number)[],
): unknown[] {
  const conditions: unknown[] = [];
  for (let i = 2; i < rowsPath.length; i += 2) {
    conditions.push(doc.getIn([...rowsPath.slice(0, i), "if"]));
  }
  return conditions;
}

/**
 * Move a component item between two rows lists (the top-level rows, a tab's or a
 * container's), identified by their YAML paths. Used for cross-container drags (e.g.
 * dragging a widget from the free canvas into a tab). The item is removed from the source
 * and inserted into the destination:
 *   - if `dest.col` is a number and `dest.row` points at an existing destination row with
 *     room, the item joins that row at that column (e.g. dropping to the left of a widget);
 *   - otherwise it becomes a new row inserted at `dest.row` (or appended).
 *
 * The item keeps the visibility of what it leaves: the conditions of its source row and of
 * the blocks holding it that the destination doesn't impose too are added to its own, so
 * content hidden from some viewers doesn't become visible to them.
 *
 * Node references for both lists are resolved up front, so removing the source row never
 * invalidates the destination even when the source sits above the destination block.
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
  let itemNode = sourceItems.items[source.col];
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

  const destConditions = [
    ...ancestorConditions(doc, dest.rowsPath),
    joinExistingRow && isMap(destRowNode) ? destRowNode.get("if") : undefined,
  ].map(conditionText);
  const carried = [
    ...ancestorConditions(doc, source.rowsPath),
    sourceRow.get("if"),
  ].filter((condition) => {
    const text = conditionText(condition);
    return text !== "" && !destConditions.includes(text);
  });
  const item =
    carried.length > 0 && ownMapAt(doc, sourceItems.items, source.col);
  if (item) {
    itemNode = item;
    const condition = combineConditions(...carried, item.get("if"));
    setInMapBeforeContent(doc, item, "if", condition);
  }

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
    // A new block's rows are an empty flow list (`rows: []`); write the rows in it in
    // block style, like any other rows.
    destSeq.flow = false;
  }

  return true;
}

/**
 * The map at the index of a list, made the list's own so that a condition written into it
 * reaches nothing else: an alias is replaced with a copy of the map it stands for, and the
 * aliases of an anchored map are replaced with copies of it. Writing into a shared map would
 * hide its other copies too, and skipping an alias would show its content to viewers it
 * should be hidden from. Undefined if the entry is not a map.
 */
function ownMapAt(
  doc: Document,
  list: unknown[],
  index: number,
): YAMLMap | undefined {
  const entry = list[index];
  if (isAlias(entry)) {
    const source = entry.resolve(doc);
    if (!isMap(source)) return undefined;
    const copy = copyWithoutAnchors(source);
    list[index] = copy;
    return copy;
  }
  if (!isMap(entry)) return undefined;
  if (entry.anchor !== undefined) {
    visit(doc, {
      Alias: (_, alias) =>
        alias.resolve(doc) === entry ? copyWithoutAnchors(entry) : undefined,
    });
    entry.anchor = undefined;
  }
  return entry;
}

/** A deep copy of the map with no anchors, which stay on the original alone. */
function copyWithoutAnchors(map: YAMLMap): YAMLMap {
  const copy = map.clone() as YAMLMap;
  visit(copy, {
    Collection: (_, node) => {
      node.anchor = undefined;
    },
    Scalar: (_, node) => {
      node.anchor = undefined;
    },
  });
  return copy;
}

/**
 * Combines `if` conditions with AND, skipping empty ones. Returns undefined if none are set.
 * Each condition renders to a boolean expression, so the combination is a valid condition too.
 */
export function combineConditions(
  ...conditions: unknown[]
): string | undefined {
  const parts = conditions.map(conditionText).filter((c) => c !== "");
  if (parts.length === 0) return undefined;
  if (parts.length === 1) return parts[0];
  return parts.map((p) => `(${p})`).join(" AND ");
}

/** The text of an `if` value, or "" when it is unset or not a condition. */
function conditionText(condition: unknown): string {
  return typeof condition === "string" || typeof condition === "boolean"
    ? String(condition).trim()
    : "";
}
