import { isSeq, type Document } from "yaml";
import type { YAMLPath } from "../layout-util";
import { defaultBlockName, setKeyBeforeContent } from "../stores/block-edit";
import { specUsesField } from "./apply-fields";
import { ControlKind } from "./controls";

/**
 * Edits of the `controls` lists in a canvas YAML document, for the visual editor.
 * A list lives on a level: the canvas (path []), a container or a tab group (["rows", i]),
 * a tab ([..., "tabs", t]) or an item ([..., "items", c]).
 */

/** The kinds of control the editor can create and edit. Slot controls are YAML-only. */
export type EditableControlKind =
  | typeof ControlKind.Measure
  | typeof ControlKind.Dimension
  | typeof ControlKind.Filter
  | typeof ControlKind.TimeRange
  | typeof ControlKind.TimeGrain;

/** A control as the editor's form holds it. */
export interface ControlDraft {
  kind: EditableControlKind;
  /** The measure or dimension the control swaps or filters. Unused for time controls. */
  field: string;
  options: string[];
  /** Default values. Unused for swaps, whose default is `field`. */
  default: string[];
  multiple: boolean;
  /** "" lets the viewer UI choose between buttons and a dropdown. */
  display: "" | "buttons" | "dropdown";
  label: string;
  /** "" uses the default name (see defaultControlName). */
  name: string;
}

export type ControlLevelKind =
  | "item"
  | "tab"
  | "group"
  | "container"
  | "canvas";

/** A level that can hold controls for an item, with the path of its YAML map. */
export interface ControlLevelRef {
  kind: ControlLevelKind;
  path: YAMLPath;
  /** The level's name or, for a tab, its label. Empty for the canvas and unnamed levels. */
  name: string;
}

export const TIME_GRAIN_OPTIONS = [
  "minute",
  "hour",
  "day",
  "week",
  "month",
  "quarter",
  "year",
];

// Keys of a control's YAML that the editor's form owns. Other keys (such as `fields`) are kept.
const MANAGED_KEYS = [
  "measure",
  "dimension",
  "filter",
  "time",
  "name",
  "label",
  "options",
  "default",
  "multiple",
  "display",
];

const NAME_REGEX = /^[a-zA-Z_][a-zA-Z0-9_]*$/;

type RawControl = Record<string, unknown>;

/** The controls declared on a level, as plain objects. */
export function readControls(
  doc: Document | undefined,
  path: YAMLPath,
): RawControl[] {
  const node = doc?.getIn([...path, "controls"]);
  if (!isSeq(node)) return [];
  const list: unknown = node.toJSON();
  return Array.isArray(list)
    ? list.map((c: unknown) => (isRecord(c) ? c : {}))
    : [];
}

/** The form values for a control's YAML, or null for a control the form can't edit (a slot control). */
export function draftFromYAML(raw: RawControl): ControlDraft | null {
  let kind: EditableControlKind;
  let field = "";
  if (typeof raw.measure === "string") {
    kind = ControlKind.Measure;
    field = raw.measure;
  } else if (typeof raw.dimension === "string") {
    kind = ControlKind.Dimension;
    field = raw.dimension;
  } else if (typeof raw.filter === "string") {
    kind = ControlKind.Filter;
    field = raw.filter;
  } else if (raw.time === "range") {
    kind = ControlKind.TimeRange;
  } else if (raw.time === "grain") {
    kind = ControlKind.TimeGrain;
  } else {
    return null;
  }
  return {
    kind,
    field,
    options: stringList(raw.options),
    default: stringList(raw.default),
    multiple: raw.multiple === true,
    display:
      raw.display === "buttons" || raw.display === "dropdown"
        ? raw.display
        : "",
    label: typeof raw.label === "string" ? raw.label : "",
    name: typeof raw.name === "string" ? raw.name : "",
  };
}

/** The YAML for a draft. Keys of `existing` that the form doesn't own are kept. */
export function controlToYAML(
  draft: ControlDraft,
  existing: RawControl = {},
): RawControl {
  const res: RawControl = {};
  const isSwap =
    draft.kind === ControlKind.Measure || draft.kind === ControlKind.Dimension;
  if (draft.kind === ControlKind.TimeRange) res.time = "range";
  else if (draft.kind === ControlKind.TimeGrain) res.time = "grain";
  else res[draft.kind] = draft.field;

  if (draft.name) res.name = draft.name;
  if (draft.label) res.label = draft.label;

  // A swap always offers the field the items show by default.
  const options =
    isSwap && !draft.options.includes(draft.field)
      ? [draft.field, ...draft.options]
      : draft.options;
  if (options.length) res.options = options;

  const multiple = draft.kind === ControlKind.Filter && draft.multiple;
  // A swap's default is its field; other defaults must be among the options, if there are any.
  const defaults = isSwap
    ? []
    : draft.default.filter((d) => !options.length || options.includes(d));
  if (defaults.length) {
    res.default = multiple ? defaults : defaults[0];
  }
  if (multiple) res.multiple = true;
  if (draft.display) res.display = draft.display;

  for (const [key, value] of Object.entries(existing)) {
    if (!MANAGED_KEYS.includes(key) && key !== "slot") res[key] = value;
  }
  return res;
}

/** The name a control has when its YAML doesn't set one. Mirrors parseCanvasControl in the runtime. */
export function defaultControlName(raw: RawControl): string {
  if (typeof raw.name === "string" && raw.name) return raw.name;
  if (raw.time === "range") return "time_range";
  if (raw.time === "grain") return "time_grain";
  if (typeof raw.slot === "string") return raw.slot.replaceAll(".", "_");
  for (const key of ["measure", "dimension", "filter"]) {
    const value = raw[key];
    if (typeof value === "string") return value;
  }
  return "";
}

/**
 * Adds a control to a level. An item gets a `name` if it has none, since its controls are
 * addressed by it. Returns the control's index in the level's list.
 */
export function addControl(
  doc: Document,
  path: YAMLPath,
  draft: ControlDraft,
): number {
  if (isItemPath(path)) ensureItemName(doc, path);
  const controls = readControls(doc, path);
  const control = withUniqueName(doc, path, controlToYAML(draft), controls);
  setKeyBeforeContent(
    doc,
    path,
    "controls",
    doc.createNode([...controls, control]),
  );
  return controls.length;
}

/** Replaces the control at an index of a level's list with the draft. */
export function updateControl(
  doc: Document,
  path: YAMLPath,
  index: number,
  draft: ControlDraft,
): void {
  const controls = readControls(doc, path);
  if (!controls[index]) return;
  const others = controls.filter((_, i) => i !== index);
  controls[index] = withUniqueName(
    doc,
    path,
    controlToYAML(draft, controls[index]),
    others,
  );
  doc.setIn([...path, "controls"], doc.createNode(controls));
}

/** Removes the control at an index. The `controls` key goes with the last control. */
export function removeControl(
  doc: Document,
  path: YAMLPath,
  index: number,
): void {
  const controls = readControls(doc, path);
  if (!controls[index]) return;
  if (controls.length === 1) {
    doc.deleteIn([...path, "controls"]);
  } else {
    doc.deleteIn([...path, "controls", index]);
  }
}

/** Moves a control one place up (-1) or down (1) in its level's list, which is the order viewers see. */
export function moveControl(
  doc: Document,
  path: YAMLPath,
  index: number,
  direction: -1 | 1,
): void {
  const node = doc.getIn([...path, "controls"]);
  const to = index + direction;
  if (!isSeq(node) || !node.items[index] || !node.items[to]) return;
  const [moved] = node.items.splice(index, 1);
  node.items.splice(to, 0, moved);
}

/**
 * Moves a control to another level, e.g. from a chart to its container, keeping its settings.
 * Returns the control's index at its new level.
 */
export function moveControlToLevel(
  doc: Document,
  from: YAMLPath,
  index: number,
  to: YAMLPath,
  draft: ControlDraft,
): number {
  const existing = readControls(doc, from)[index] ?? {};
  // Remove first, so the control's name is free at the new level.
  removeControl(doc, from, index);
  if (isItemPath(to)) ensureItemName(doc, to);
  const controls = readControls(doc, to);
  const control = withUniqueName(
    doc,
    to,
    controlToYAML(draft, existing),
    controls,
  );
  setKeyBeforeContent(
    doc,
    to,
    "controls",
    doc.createNode([...controls, control]),
  );
  return controls.length;
}

/**
 * The levels whose controls apply to the item at a path, nearest first: the item itself,
 * then its tab, tab group or container, then the canvas.
 */
export function controlLevelsForItem(
  doc: Document | undefined,
  itemPath: YAMLPath,
): ControlLevelRef[] {
  const levels: ControlLevelRef[] = [{ kind: "canvas", path: [], name: "" }];
  const text = (path: YAMLPath) => {
    const value: unknown = doc?.getIn(path);
    return typeof value === "string" ? value : "";
  };

  let i = 0;
  while (itemPath[i] === "rows") {
    const rowPath = itemPath.slice(0, i + 2);
    const next = itemPath[i + 2];
    if (next === "tabs") {
      const tabPath = itemPath.slice(0, i + 4);
      levels.push({
        kind: "group",
        path: rowPath,
        name:
          text([...rowPath, "name"]) ||
          defaultBlockName("tab-group", Number(rowPath[1])),
      });
      levels.push({
        kind: "tab",
        path: tabPath,
        name: text([...tabPath, "label"]) || text([...tabPath, "name"]),
      });
      i += 4;
    } else if (next === "rows") {
      levels.push({
        kind: "container",
        path: rowPath,
        name:
          text([...rowPath, "name"]) ||
          defaultBlockName("container", Number(rowPath[1])),
      });
      i += 2;
    } else {
      if (next === "items") {
        const path = itemPath.slice(0, i + 4);
        levels.push({ kind: "item", path, name: text([...path, "name"]) });
      }
      break;
    }
  }
  return levels.reverse();
}

/**
 * How many items under a level show the field, i.e. how many a swap of it would change.
 * Counts items that opt out with `ignore_controls: true` as unaffected.
 */
export function countItemsUsingField(
  doc: Document | undefined,
  path: YAMLPath,
  field: string,
): number {
  const root = toJSON(path.length ? doc?.getIn(path) : doc);
  if (!isRecord(root) || !field) return 0;

  const own = isItemPath(path);
  let count = 0;
  const visitItem = (item: unknown) => {
    if (!isRecord(item)) return;
    if (!own && item.ignore_controls === true) return;
    const uses = Object.entries(item).some(
      ([key, value]) =>
        key !== "controls" && isRecord(value) && specUsesField(value, field),
    );
    if (uses) count++;
  };
  const visitRows = (rows: unknown) => {
    if (!Array.isArray(rows)) return;
    for (const row of rows) {
      if (!isRecord(row)) continue;
      if (Array.isArray(row.items)) row.items.forEach(visitItem);
      if (Array.isArray(row.tabs)) {
        for (const tab of row.tabs) {
          if (isRecord(tab)) visitRows(tab.rows);
        }
      }
      visitRows(row.rows);
    }
  };

  if (own) {
    visitItem(root);
  } else if (Array.isArray(root.tabs)) {
    visitRows([root]);
  } else {
    visitRows(root.rows);
  }
  return count;
}

export function isItemPath(path: YAMLPath): boolean {
  return path[path.length - 2] === "items";
}

/** Gives the item at a path a `name` if it has none: its component type, numbered if taken. */
function ensureItemName(doc: Document, path: YAMLPath): void {
  const current: unknown = doc.getIn([...path, "name"]);
  if (typeof current === "string" && current) return;

  const item = toJSON(doc.getIn(path));
  const type = isRecord(item)
    ? Object.keys(item).find((key) => key !== "controls" && isRecord(item[key]))
    : undefined;
  const base = type && NAME_REGEX.test(type) ? type : "widget";
  setKeyBeforeContent(doc, path, "name", uniqueName(base, takenNames(doc)));
}

/**
 * Sets `name` on a control whose name is taken by another control of its level or, for a
 * canvas control, by an item or block (they share the top level of the URL keys).
 */
function withUniqueName(
  doc: Document,
  path: YAMLPath,
  control: RawControl,
  siblings: RawControl[],
): RawControl {
  const taken = new Set(siblings.map(defaultControlName));
  if (path.length === 0) {
    for (const name of takenNames(doc, false)) taken.add(name);
  }
  const name = defaultControlName(control);
  if (!taken.has(name)) return control;

  // Insert the name after the key that says what the control changes.
  const [[targetKey, target], ...rest] = Object.entries(control).filter(
    ([key]) => key !== "name",
  );
  const kindSuffix =
    typeof control.filter === "string" ? `${name}_filter` : name;
  const unique = taken.has(kindSuffix) ? uniqueName(name, taken) : kindSuffix;
  return { [targetKey]: target, name: unique, ...Object.fromEntries(rest) };
}

/** Names of the canvas's items and blocks and, optionally, its canvas-level controls. */
function takenNames(doc: Document, withCanvasControls = true): Set<string> {
  const names = new Set<string>();
  const root = toJSON(doc);
  if (!isRecord(root)) return names;
  if (withCanvasControls && Array.isArray(root.controls)) {
    for (const c of root.controls) {
      if (isRecord(c)) names.add(defaultControlName(c));
    }
  }
  const visitRows = (rows: unknown) => {
    if (!Array.isArray(rows)) return;
    for (const row of rows) {
      if (!isRecord(row)) continue;
      if (typeof row.name === "string") names.add(row.name);
      if (Array.isArray(row.items)) {
        for (const item of row.items) {
          if (isRecord(item) && typeof item.name === "string") {
            names.add(item.name);
          }
        }
      }
      if (Array.isArray(row.tabs)) {
        for (const tab of row.tabs) {
          if (isRecord(tab)) visitRows(tab.rows);
        }
      }
      visitRows(row.rows);
    }
  };
  visitRows(root.rows);
  return names;
}

function uniqueName(base: string, taken: Set<string>): string {
  if (!taken.has(base)) return base;
  let n = 2;
  while (taken.has(`${base}_${n}`)) n++;
  return `${base}_${n}`;
}

function toJSON(node: unknown): unknown {
  if (node && typeof node === "object" && "toJSON" in node) {
    return (node as { toJSON: () => unknown }).toJSON();
  }
  return node;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

function stringList(value: unknown): string[] {
  if (Array.isArray(value)) return value.map(String);
  if (value === undefined || value === null || value === "") return [];
  return [String(value)];
}
