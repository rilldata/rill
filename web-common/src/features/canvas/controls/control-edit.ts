import { bucketYamlRanges } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import type {
  V1ExploreTimeRange,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { Duration } from "luxon";
import { isSeq, type Document } from "yaml";
import type { ComponentInputParam } from "../inspector/types";
import type { YAMLPath } from "../layout-util";
import { defaultBlockName, setKeyBeforeContent } from "../stores/block-edit";
import { specUsesField, swapRepeatsField } from "./apply-fields";
import { ControlKind } from "./controls";

/**
 * Edits of the `controls` lists in a canvas YAML document, for the visual editor.
 * A list lives on a level: the canvas (path []), a container or a tab group (["rows", i]),
 * a tab ([..., "tabs", t]) or an item ([..., "items", c]).
 */

/** The kinds of control the editor can create and edit. */
export type EditableControlKind =
  | typeof ControlKind.Measure
  | typeof ControlKind.Dimension
  | typeof ControlKind.Slot
  | typeof ControlKind.Filter
  | typeof ControlKind.TimeRange
  | typeof ControlKind.TimeGrain;

/** A control as the editor's form holds it. */
export interface ControlDraft {
  kind: EditableControlKind;
  /**
   * The measure or dimension the control swaps or filters, or for a slot control the widget's
   * property whose fields viewers pick (e.g. `measures`). Unused for time controls.
   */
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

/** A widget property whose fields a slot control lets viewers pick, e.g. a KPI grid's `measures`. */
export interface SlotTarget {
  target: string;
  label: string;
  /** The kinds of field the property holds. */
  allowed: ("measure" | "dimension")[];
}

/**
 * The properties of a widget that hold a list of measures or dimensions (its inspector inputs
 * that pick several fields), which viewers can pick with a slot control.
 */
export function listFieldTargets(
  inputs: Partial<Record<string, ComponentInputParam>>,
): SlotTarget[] {
  return Object.entries(inputs).flatMap(([target, config]) => {
    if (
      !config ||
      (config.type !== "multi_fields" && config.type !== "multi_fields_format")
    ) {
      return [];
    }
    const allowed = (config.meta?.allowedTypes ?? []).filter(
      (t): t is "measure" | "dimension" => t === "measure" || t === "dimension",
    );
    return allowed.length
      ? [{ target, label: config.label ?? target, allowed }]
      : [];
  });
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

/**
 * The time ranges the canvas's filter bar offers, written as its menu writes them: the canvas's
 * `time_ranges`, or the default ranges for the metrics views' smallest time grain.
 */
export function filterBarTimeRanges(
  timeRanges: V1ExploreTimeRange[],
  minTimeGrain: V1TimeGrain | undefined,
  maxQueryTimeRange?: Duration,
): string[] {
  const buckets = bucketYamlRanges(
    timeRanges,
    minTimeGrain,
    true,
    maxQueryTimeRange,
  );
  const ranges = [
    ...buckets.latest,
    ...buckets.periodToDate,
    ...buckets.previous,
    ...buckets.custom,
  ].map((r) => (r.asOfLabel ? r.toString() : r.interval.toString()));
  return buckets.allTime ? [...ranges, "inf"] : ranges;
}

// Keys of a control's YAML that the editor's form owns. Other keys (such as `fields`) are kept.
const MANAGED_KEYS = [
  "measure",
  "dimension",
  "slot",
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

/** The form values for a control's YAML, or null for a control the form can't edit. */
export function draftFromYAML(raw: RawControl): ControlDraft | null {
  let kind: EditableControlKind;
  let field = "";
  if (typeof raw.measure === "string") {
    kind = ControlKind.Measure;
    field = raw.measure;
  } else if (typeof raw.dimension === "string") {
    kind = ControlKind.Dimension;
    field = raw.dimension;
  } else if (typeof raw.slot === "string") {
    kind = ControlKind.Slot;
    field = raw.slot;
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

  const multiple =
    (draft.kind === ControlKind.Filter || draft.kind === ControlKind.Slot) &&
    draft.multiple;
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
    if (!MANAGED_KEYS.includes(key)) res[key] = value;
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
 * Counts items that opt out of the control with `ignore_controls`, or that a nearer control
 * of the same kind and field overrides, as unaffected.
 */
export function countItemsUsingField(
  doc: Document | undefined,
  path: YAMLPath,
  kind: EditableControlKind,
  field: string,
  controlName = field,
): number {
  if (!field) return 0;
  return listAffectedItems(doc, path, kind, field, controlName, (spec) =>
    specUsesField(spec, field),
  ).length;
}

/** An item a control would change: its name, and the title viewers see if it has one. */
export interface AffectedItem {
  name: string;
  title: string;
}

/**
 * The items under a level that a control would change, in canvas order: those `applies` accepts
 * a component spec of, less the ones that opt out with `ignore_controls` or that a nearer
 * control of the same kind and field reaches. `applies` decides per kind, e.g. a swap needs the
 * field in the spec and a filter needs the dimension in the spec's metrics view.
 */
export function listAffectedItems(
  doc: Document | undefined,
  path: YAMLPath,
  kind: EditableControlKind,
  field: string,
  controlName: string,
  applies: (spec: Record<string, unknown>) => boolean,
): AffectedItem[] {
  return itemsUnderLevel(doc, path, kind, field, controlName)
    .filter(({ specs }) => specs.some(applies))
    .map(({ name, specs }) => ({ name, title: itemTitle(name, specs) }));
}

/** The title viewers see on an item, or its name if it has none. */
function itemTitle(name: string, specs: Record<string, unknown>[]): string {
  const title = specs.find((spec) => typeof spec.title === "string")?.title as
    | string
    | undefined;
  return title?.trim() || name;
}

/**
 * The name of another control on the level that changes the same thing as one of this kind and
 * field, or "" if there is none. Two such controls on one level compete and only the first
 * applies, so the runtime rejects the second; `exceptName` is the control being edited.
 */
export function findLevelConflict(
  doc: Document | undefined,
  path: YAMLPath,
  kind: EditableControlKind,
  field: string,
  exceptName = "",
): string {
  const [key, value] = controlTargetEntry(kind, field);
  const conflict = readControls(doc, path).find(
    (c) => c[key] === value && defaultControlName(c) !== exceptName,
  );
  return conflict ? defaultControlName(conflict) : "";
}

/**
 * A test of whether the level already has a control that changes what one of this kind and field
 * would, leaving out the control named `exceptName` (see `findLevelConflict`). The editor offers
 * only the kinds, fields and levels that are free, rather than a choice it would then reject.
 */
export function takenOnLevel(
  doc: Document | undefined,
  path: YAMLPath,
  exceptName = "",
): (kind: EditableControlKind, field: string) => boolean {
  const taken = new Set<string>();
  for (const c of readControls(doc, path)) {
    if (defaultControlName(c) === exceptName) continue;
    for (const key of ["time", "measure", "dimension", "slot", "filter"]) {
      if (typeof c[key] === "string") taken.add(`${key}:${c[key]}`);
    }
  }
  return (kind, field) => taken.has(controlTargetEntry(kind, field).join(":"));
}

/** The YAML key and value that say what a control of this kind changes, e.g. `time: range`. */
function controlTargetEntry(
  kind: EditableControlKind,
  field: string,
): [string, string] {
  if (kind === ControlKind.TimeRange) return ["time", "range"];
  if (kind === ControlKind.TimeGrain) return ["time", "grain"];
  return [kind, field];
}

/**
 * The first option of a swap that an item under the level already shows next to the field, so
 * choosing it would show that field twice there. The item is named by its title, if it has one. The runtime reports such a control and leaves it
 * out, so the editor checks before writing it.
 */
export function findRepeatedOption(
  doc: Document | undefined,
  path: YAMLPath,
  kind: EditableControlKind,
  field: string,
  options: string[],
  controlName = field,
): { option: string; item: string } | null {
  if (!field) return null;
  for (const { name, specs } of itemsUnderLevel(
    doc,
    path,
    kind,
    field,
    controlName,
  )) {
    for (const spec of specs) {
      if (!specUsesField(spec, field)) continue;
      const option = options.find(
        (o) => o !== field && swapRepeatsField(spec, field, o),
      );
      if (option) return { option, item: itemTitle(name, specs) };
    }
  }
  return null;
}

/**
 * The items a control declared on a level applies to: the item itself, or the items below a
 * block or the canvas, less those that opt out of the control with `ignore_controls` and those
 * a nearer control of the same kind and field reaches, since it wins (see controlsForPath).
 * Each comes with its name (or component type) and its component specs as plain objects.
 */
function itemsUnderLevel(
  doc: Document | undefined,
  path: YAMLPath,
  kind: EditableControlKind,
  field: string,
  controlName: string,
): { name: string; specs: Record<string, unknown>[] }[] {
  const root = toJSON(path.length ? doc?.getIn(path) : doc);
  if (!isRecord(root)) return [];

  // The names of a level's controls of the same kind and field.
  const [targetKey, target] = controlTargetEntry(kind, field);
  const sameControls = (level: Record<string, unknown>) =>
    Array.isArray(level.controls)
      ? level.controls
          .filter(
            (c): c is RawControl => isRecord(c) && c[targetKey] === target,
          )
          .map(defaultControlName)
      : [];
  const own = isItemPath(path);
  const res: { name: string; specs: Record<string, unknown>[] }[] = [];
  // `nearer` holds the names of such controls on the levels between this one and the item.
  const visitItem = (item: unknown, nearer: string[]) => {
    if (!isRecord(item)) return;
    const ignore = item.ignore_controls;
    const ignores = (name: string) =>
      ignore === true || (Array.isArray(ignore) && ignore.includes(name));
    // The item's own control wins, and so does a nearer one it doesn't ignore.
    if (
      !own &&
      (sameControls(item).length > 0 ||
        ignores(controlName) ||
        nearer.some((name) => !ignores(name)))
    ) {
      return;
    }
    const entries = Object.entries(item).filter(
      ([key, value]) => key !== "controls" && isRecord(value),
    );
    res.push({
      name: typeof item.name === "string" ? item.name : (entries[0]?.[0] ?? ""),
      specs: entries.map(([, value]) => value as Record<string, unknown>),
    });
  };
  const visitTabs = (tabs: unknown, nearer: string[]) => {
    if (!Array.isArray(tabs)) return;
    for (const tab of tabs) {
      if (isRecord(tab)) visitRows(tab.rows, [...nearer, ...sameControls(tab)]);
    }
  };
  const visitRows = (rows: unknown, nearer: string[]) => {
    if (!Array.isArray(rows)) return;
    for (const row of rows) {
      if (!isRecord(row)) continue;
      const inner = [...nearer, ...sameControls(row)];
      if (Array.isArray(row.items)) {
        for (const item of row.items) visitItem(item, inner);
      }
      visitTabs(row.tabs, inner);
      visitRows(row.rows, inner);
    }
  };

  if (own) {
    visitItem(root, []);
  } else if (Array.isArray(root.tabs)) {
    visitTabs(root.tabs, []);
  } else {
    visitRows(root.rows, []);
  }
  return res;
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
