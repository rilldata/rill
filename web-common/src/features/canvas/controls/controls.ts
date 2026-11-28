import { DateTimeUnitToV1TimeGrain } from "@rilldata/web-common/lib/time/new-grains";
import type {
  V1CanvasControl,
  V1CanvasRow,
  V1CanvasSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";

/**
 * Viewer controls of a canvas: which controls apply to which item, and how their values are
 * stored in the URL. A control is declared on the canvas, a container, a tab group, a tab or
 * an item, and applies to the item it is declared on or to every item nested under its level.
 *
 * The runtime implements the same scope rules in runtime/canvas/controls.go
 * (ResolveItemControls); keep them in sync.
 */

export const ControlKind = {
  Measure: "measure",
  Dimension: "dimension",
  Slot: "slot",
  Filter: "filter",
  TimeRange: "time_range",
  TimeGrain: "time_grain",
} as const;

/** URL params that hold control values start with this prefix. */
export const CONTROL_URL_PREFIX = "c.";

export type ControlLevel = "canvas" | "container" | "group" | "tab" | "item";

export interface ResolvedControl {
  control: V1CanvasControl;
  level: ControlLevel;
  /** Names of the level the control is declared on, outermost first. Empty for the canvas. */
  scope: string[];
  /** The control's URL param, e.g. `c.deep_dive.publisher`. */
  urlKey: string;
  /** What the control changes. Controls with the same target compete; the nearest wins. */
  targetKey: string;
}

/** A field name, time setting or raw primitive dimension value. */
export type ControlValue = string | number | boolean | null;

/** A control in effect for an item (null values mean its default applies). */
export interface AppliedControl extends ResolvedControl {
  values: ControlValue[] | null;
  /**
   * For a slot control: the fields the item's YAML sets there, which viewers see and start from
   * while the control is on its default.
   */
  shownValues?: ControlValue[];
  /** True when the embedding page fixed the control's value: viewers don't see or change it. */
  locked?: boolean;
}

/**
 * Values fixed from outside the canvas (an embed's `locked_controls`), keyed by the control's
 * URL key with or without the `c.` prefix, in the URL value format.
 */
export type LockedControls = Record<string, string>;

/** The fixed value for a control, or null if it isn't locked. */
export function lockedControlValue(
  locked: LockedControls,
  urlKey: string,
): string | null {
  const key = Object.hasOwn(locked, urlKey)
    ? urlKey
    : urlKey.slice(CONTROL_URL_PREFIX.length);
  return Object.hasOwn(locked, key) ? locked[key] : null;
}

/** Encodes one segment of a control's URL key, so a "." in a name isn't read as a separator. */
export function encodeControlKeyPart(name: string): string {
  return encodeURIComponent(name).replace(/\./g, "%2E");
}

export function controlUrlKey(scope: string[], controlName: string): string {
  return (
    CONTROL_URL_PREFIX +
    [...scope, controlName].map(encodeControlKeyPart).join(".")
  );
}

function resolve(
  control: V1CanvasControl,
  level: ControlLevel,
  scope: string[],
  singleSelect: ReadonlySet<string>,
): ResolvedControl {
  return {
    control: followSingleSelect(control, singleSelect),
    level,
    scope,
    urlKey: controlUrlKey(scope, control.name ?? ""),
    targetKey: `${control.kind}:${control.target ?? ""}`,
  };
}

function resolveAll(
  controls: V1CanvasControl[] | undefined,
  level: ControlLevel,
  scope: string[],
  singleSelect: ReadonlySet<string>,
): ResolvedControl[] {
  return (controls ?? []).map((c) => resolve(c, level, scope, singleSelect));
}

/**
 * A filter control on a dimension the canvas lists under `filters.single_select` holds one value,
 * whatever its own `multiple` says, so it follows that setting as the filter bar does. Its default
 * keeps the last value, as the filter bar chip narrows a filter.
 */
function followSingleSelect(
  control: V1CanvasControl,
  singleSelect: ReadonlySet<string>,
): V1CanvasControl {
  if (control.kind !== ControlKind.Filter || !control.multiple) return control;
  const dimensions = [
    control.target ?? "",
    ...Object.values(control.fields ?? {}),
  ];
  if (!dimensions.some((d) => singleSelect.has(d))) return control;
  return { ...control, multiple: false, default: control.default?.slice(-1) };
}

/** Every control declared anywhere on the canvas. */
export function collectControls(
  spec: V1CanvasSpec | undefined,
): ResolvedControl[] {
  if (!spec) return [];
  const singleSelect = new Set(spec.singleSelectFilters ?? []);
  const res = resolveAll(spec.controls, "canvas", [], singleSelect);
  const walk = (rows: V1CanvasRow[] | undefined) => {
    for (const row of rows ?? []) {
      const group = row.tabGroup;
      if (group) {
        const groupName = group.name ?? "";
        res.push(
          ...resolveAll(group.controls, "group", [groupName], singleSelect),
        );
        for (const tab of group.tabs ?? []) {
          res.push(
            ...resolveAll(
              tab.controls,
              "tab",
              [groupName, tab.name ?? ""],
              singleSelect,
            ),
          );
          walk(tab.rows);
        }
      }
      const container = row.container;
      if (container) {
        res.push(
          ...resolveAll(
            container.controls,
            "container",
            [container.name ?? ""],
            singleSelect,
          ),
        );
        walk(container.rows);
      }
      for (const item of row.items ?? []) {
        if (item.name) {
          res.push(
            ...resolveAll(item.controls, "item", [item.name], singleSelect),
          );
        }
      }
    }
  };
  walk(spec.rows);
  return res;
}

/**
 * The controls in effect for the item at the given YAML path
 * (e.g. ["rows", 2, "tabs", 0, "rows", 1, "items", 0, "bar_chart"]), nearest level first.
 * For each target only the nearest control is kept, and the item's `ignore_controls`
 * drops controls of the enclosing levels (never its own).
 */
export function controlsForPath(
  spec: V1CanvasSpec | undefined,
  path: (string | number)[],
): ResolvedControl[] {
  if (!spec || path[0] !== "rows") return [];
  const singleSelect = new Set(spec.singleSelectFilters ?? []);

  // Levels enclosing the item, outermost first.
  const enclosing: ResolvedControl[][] = [
    resolveAll(spec.controls, "canvas", [], singleSelect),
  ];
  let rows = spec.rows ?? [];
  let i = 1;
  for (;;) {
    const row = rows[Number(path[i])];
    if (!row) return [];
    const next = path[i + 1];
    if (next === "tabs") {
      const group = row.tabGroup;
      const tab = group?.tabs?.[Number(path[i + 2])];
      if (!group || !tab) return [];
      const groupName = group.name ?? "";
      enclosing.push(
        resolveAll(group.controls, "group", [groupName], singleSelect),
      );
      enclosing.push(
        resolveAll(
          tab.controls,
          "tab",
          [groupName, tab.name ?? ""],
          singleSelect,
        ),
      );
      rows = tab.rows ?? [];
      i += 4; // row index, "tabs", tab index, "rows"
      continue;
    }
    if (next === "rows") {
      const container = row.container;
      if (!container) return [];
      enclosing.push(
        resolveAll(
          container.controls,
          "container",
          [container.name ?? ""],
          singleSelect,
        ),
      );
      rows = container.rows ?? [];
      i += 2; // row index, "rows"
      continue;
    }
    if (next !== "items") return [];
    const item = row.items?.[Number(path[i + 2])];
    if (!item) return [];

    const res: ResolvedControl[] = [];
    const seen = new Set<string>();
    const add = (c: ResolvedControl) => {
      if (seen.has(c.targetKey)) return;
      seen.add(c.targetKey);
      res.push(c);
    };
    if (item.name) {
      resolveAll(item.controls, "item", [item.name], singleSelect).forEach(add);
    }
    if (!item.ignoreAllControls) {
      const ignored = new Set(item.ignoreControls ?? []);
      for (const level of enclosing.reverse()) {
        for (const c of level) {
          if (!ignored.has(c.control.name ?? "")) add(c);
        }
      }
    }
    return res;
  }
}

/** The values a control has when the URL doesn't set it. Null means the items keep what their YAML says. */
export function defaultControlValues(
  control: V1CanvasControl,
): string[] | null {
  switch (control.kind) {
    case ControlKind.Measure:
    case ControlKind.Dimension:
      return [control.target ?? ""];
    case ControlKind.Slot:
      return null;
    case ControlKind.Filter:
      return control.default ?? [];
    default:
      return control.default?.length ? control.default : null;
  }
}

/**
 * How an empty-string value is written in a control's URL value, e.g. a dimension's blank
 * value. An empty param already means "cleared", so the empty string needs its own token:
 * two double quotes, which encodeURIComponent never produces for another value.
 */
const EMPTY_VALUE_TOKEN = '""';

/**
 * Parses a control's URL value. A missing param (null) means the default; an empty param means
 * the viewer cleared the control. Several values are comma-separated, each URI-encoded, with
 * the empty string written as `""`.
 * A value that the control doesn't allow is invalid, and the default applies instead.
 */
export function parseControlValue(
  control: V1CanvasControl,
  raw: string | null,
): { values: ControlValue[] | null; valid: boolean } {
  if (raw === null)
    return { values: defaultControlValues(control), valid: true };

  let values: ControlValue[];
  try {
    values =
      raw === ""
        ? []
        : raw.split(",").map((v) => {
            if (v === EMPTY_VALUE_TOKEN) return "";
            // A colon is always escaped in string values, so it can tag typed filter values.
            if (control.kind === ControlKind.Filter && v.startsWith(":")) {
              const value: unknown = JSON.parse(decodeURIComponent(v.slice(1)));
              if (
                value === null ||
                typeof value === "boolean" ||
                (typeof value === "number" && Number.isFinite(value))
              )
                return value;
              throw new Error("Invalid filter value");
            }
            return decodeURIComponent(v);
          });
  } catch {
    return { values: defaultControlValues(control), valid: false };
  }

  // A filter that holds one value keeps the last of several, as a filter bar chip does once its
  // dimension is single-select, e.g. after `filters.single_select` lists it.
  if (
    control.kind === ControlKind.Filter &&
    !control.multiple &&
    values.length > 1
  ) {
    values = values.slice(-1);
  }

  const options = control.options ?? [];
  const allowed =
    options.length === 0 ||
    values.every((v) => typeof v === "string" && options.includes(v));
  const countOk = control.multiple || values.length <= 1;
  // Only a filter can be cleared: the other kinds always need a value.
  const emptyOk = values.length > 0 || control.kind === ControlKind.Filter;
  if (!allowed || !countOk || !emptyOk) {
    return { values: defaultControlValues(control), valid: false };
  }
  return { values, valid: true };
}

/** The URL value for a control, or undefined when the values are the default and the param is omitted. */
export function serializeControlValue(
  control: V1CanvasControl,
  values: ControlValue[] | null,
): string | undefined {
  if (values === null) return undefined;
  const def = defaultControlValues(control);
  if (
    def !== null &&
    def.length === values.length &&
    def.every((v, i) => v === values[i])
  ) {
    return undefined;
  }
  return values
    .map((v) =>
      typeof v === "string"
        ? v === ""
          ? EMPTY_VALUE_TOKEN
          : encodeURIComponent(v)
        : `:${encodeURIComponent(JSON.stringify(v))}`,
    )
    .join(",");
}

/**
 * The metrics views to read a filter control's values from. A control filters its target
 * dimension, or the dimension its `fields` maps a metrics view to. Values come from the views
 * that have the target itself if any do, otherwise from the mapped ones; either way only from
 * views that agree on the dimension's name, since one query set reads one dimension.
 */
export function filterValueViews(
  control: V1CanvasControl,
  views: string[],
  hasDimension: (view: string, dimension: string) => boolean,
): string[] {
  const dimensionFor = (view: string) =>
    control.fields?.[view] ?? control.target ?? "";
  const candidates = views.filter((v) => hasDimension(v, dimensionFor(v)));
  const pick =
    candidates.find((v) => dimensionFor(v) === control.target) ?? candidates[0];
  if (pick === undefined) return [];
  return candidates.filter((v) => dimensionFor(v) === dimensionFor(pick));
}

/**
 * An identity for comparing a selected value with an option. A number written as text, as in a
 * hand-written link or `locked_controls`, matches the dimension's raw number; null stays distinct
 * from the text "null".
 */
export function controlValueKey(value: ControlValue): string {
  return value === null ? "null:" : `value:${String(value)}`;
}

/**
 * Visible controls declared on the requested levels, shared by live and PDF renders.
 * A control is left out when a nearer one of the levels (a tab, under its tab group) declares a
 * control for the same target, since that one wins for the widgets shown there.
 */
export function controlsForScopes(
  controls: AppliedControl[],
  scopes: string[][],
): AppliedControl[] {
  const keys = new Set(scopes.map((scope) => JSON.stringify(scope)));
  const shown = controls.filter(
    (c) => c.level !== "item" && !c.locked && keys.has(JSON.stringify(c.scope)),
  );
  return shown.filter(
    (c) =>
      !shown.some(
        (d) => d.targetKey === c.targetKey && d.scope.length > c.scope.length,
      ),
  );
}

/**
 * The grains among a time grain control's options that the metrics views offer, as the filter
 * bar's grain picker does: their `time_grains`, or the grains at or above their smallest time
 * grain. All of them while the metrics views aren't known.
 */
export function offeredTimeGrains(
  grains: string[],
  available: V1TimeGrain[],
): string[] {
  if (!available.length) return grains;
  return grains.filter((grain) => {
    const v1Grain = DateTimeUnitToV1TimeGrain[grain as DateTimeUnit];
    return !!v1Grain && available.includes(v1Grain);
  });
}

/** Metrics views of widgets to which each resolved control actually applies. */
export function controlMetricsViews(
  components: { metricsViewName: string; controls: AppliedControl[] }[],
): Map<string, string[]> {
  const views = new Map<string, Set<string>>();
  for (const component of components) {
    if (!component.metricsViewName) continue;
    for (const control of component.controls) {
      if (!views.has(control.urlKey)) views.set(control.urlKey, new Set());
      views.get(control.urlKey)!.add(component.metricsViewName);
    }
  }
  return new Map([...views].map(([key, names]) => [key, [...names]]));
}
