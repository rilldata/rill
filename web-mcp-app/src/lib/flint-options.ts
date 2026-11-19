import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  computeChartTypeStates,
  type ChartAssemblyInput,
  type ChartEncoding,
  type ChartOption,
} from "flint-chart/core";
import {
  getChartOptions,
  getChartTransform,
  vlGetTemplateDef,
} from "flint-chart/vegalite";
import type {
  WireField,
  WireFieldKind,
  WireFlintChartSpec,
  WireFlintEncodingValue,
} from "./contract";
import { comparisonChannel, type FlintCompileInput } from "./flint-input";

/**
 * Presentation controls derived from Flint's own option model, so the panel can never offer a
 * knob the compiler ignores: applicable binary and discrete chart properties (`getChartOptions`),
 * the alternative arrangements of the same bindings (the `arrange` control of
 * `getChartTransform`), and Flint's chart-type transitions, the sibling types that keep the
 * chart's meaning (`computeChartTypeStates`). Every edit is a new chart spec applied
 * client-side; nothing here talks to the server.
 *
 * Adapted from `packages/flint-mcp/ui/src/options.ts` in microsoft/flint-chart (MIT License,
 * Copyright (c) Microsoft Corporation), restricted to the controls this widget renders.
 */

export interface PanelChoice {
  value: string;
  label: string;
}

export interface PanelProperty {
  key: string;
  label: string;
  choices: PanelChoice[];
  /** The current choice, as `valueKey` of the property's value. */
  value: string | undefined;
}

export interface FlintPanel {
  /** The chart types the result's spec can move to; undefined when Flint offers no transition. */
  chartType: { current: string; choices: PanelChoice[] } | undefined;
  /** Alternative arrangements of the same bindings (swap axes, split into panels, ...). */
  layout:
    | {
        /** The chart property Flint reads the chosen arrangement from. */
        key: string;
        choices: PanelChoice[];
        value: string;
      }
    | undefined;
  properties: PanelProperty[];
  apply: {
    chartType(type: string): WireFlintChartSpec;
    layout(id: string): WireFlintChartSpec;
    property(key: string, choice: string): WireFlintChartSpec;
  };
}

/** The toolbar stays readable: a few properties and a few views. */
export const MAX_PANEL_PROPERTIES = 3;
export const MAX_LAYOUT_CHOICES = 4;

/** The Vega-Lite type of a field of each kind, for a binding that declares none. */
const KIND_TYPES: Record<WireFieldKind, NonNullable<ChartEncoding["type"]>> = {
  measure: "quantitative",
  dimension: "nominal",
  time: "temporal",
};

/**
 * Localized labels of the steps Flint composes its arrange states from; a state id is a path of
 * steps (`flip:x-y|augment:row`), which Flint labels in operator notation ("X ⇄ Y · Color + Rows").
 */
const LAYOUT_STEP_LABELS: Partial<Record<string, () => string>> = {
  "flip:x-y": m.mcp_chart_layout_swap_axes,
  "swap:x-color": m.mcp_chart_layout_swap_x_color,
  "swap:y-color": m.mcp_chart_layout_swap_y_color,
  "swap:x-size": m.mcp_chart_layout_swap_x_size,
  "swap:y-size": m.mcp_chart_layout_swap_y_size,
  // The series also splits the chart into small multiples.
  "augment:column": m.mcp_chart_layout_split_columns,
  "augment:row": m.mcp_chart_layout_split_rows,
  // The small multiples' field moves to another channel.
  "series:color": m.mcp_chart_layout_combine_panels,
  "series:column": m.mcp_chart_layout_panels_columns,
  "series:row": m.mcp_chart_layout_panels_rows,
};

/** Stable string key for an arbitrary option value (handles undefined and objects). */
export function valueKey(value: unknown): string {
  return JSON.stringify(value ?? null);
}

/** The label of an arrange state; a state with a step this widget does not know keeps Flint's. */
export function layoutLabel(id: string, flintLabel: string): string {
  if (id === "default") return m.mcp_chart_layout_default();
  const steps = id.split("|").map((step) => LAYOUT_STEP_LABELS[step]);
  if (!steps.every((label) => label !== undefined)) return flintLabel;
  return steps.map((label) => label()).join(" · ");
}

function assemblyInput(compile: FlintCompileInput): ChartAssemblyInput {
  return {
    data: { values: compile.rows },
    semantic_types: compile.fields.semantic_types,
    field_display_names: compile.fields.field_display_names,
    chart_spec: compile.spec as ChartAssemblyInput["chart_spec"],
    options: { addTooltips: true },
  };
}

/**
 * The binding a moved field keeps: its declared type and its Rill `limit`, so a dimension read as
 * time stays one and the top-N survives the move. Its sort keys are dropped, since `sortBy` names
 * a channel of the old chart type and the new type's own ordering applies.
 */
function encodingFor(
  chartSpec: WireFlintChartSpec,
  field: string,
): WireFlintEncodingValue {
  for (const value of Object.values(chartSpec.encodings ?? {})) {
    if (typeof value !== "object" || Array.isArray(value)) continue;
    if (value.field !== field) continue;
    const { type, limit } = value;
    if (type === undefined && limit === undefined) return field;
    return {
      field,
      ...(type !== undefined && { type }),
      ...(limit !== undefined && { limit }),
    };
  }
  return field;
}

/**
 * The spec's bindings as Flint's transitions read them: one typed encoding per channel, the type
 * from the field's kind unless the binding declares one. Undefined for a spec that binds several
 * fields on a channel (a series fold), which this widget does not move.
 */
function typedEncodings(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField>,
): Record<string, ChartEncoding> | undefined {
  const typed: Record<string, ChartEncoding> = {};
  for (const [channel, value] of Object.entries(chartSpec.encodings ?? {})) {
    if (Array.isArray(value)) return undefined;
    const encoding = typeof value === "string" ? { field: value } : value;
    if (!encoding.field) continue;
    const kind = fields[encoding.field]?.kind;
    typed[channel] = {
      field: encoding.field,
      type: encoding.type ?? (kind ? KIND_TYPES[kind] : "nominal"),
      ...(encoding.aggregate && { aggregate: encoding.aggregate }),
    };
  }
  return typed;
}

/**
 * The specs the chart can move to, keyed by chart type: Flint's one-hop transitions of the spec
 * (a line chart to an area or a bar chart, a grouped bar chart to a stacked one, ...), with the
 * fields re-routed the way Flint's own chart-type control does. Each is a plain chart spec rather
 * than Flint's `chartType` override, so that the comparison fold, the server's validation and
 * `chart_data` all see the chart type that is drawn. Chart properties are template-specific, so
 * a moved spec starts from the new type's defaults.
 */
function chartTypeMoves(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField>,
  rows: FlintCompileInput["rows"],
): Map<string, WireFlintChartSpec> {
  const moves = new Map<string, WireFlintChartSpec>();
  const template = vlGetTemplateDef(chartSpec.chartType);
  const encodings = typedEncodings(chartSpec, fields);
  if (!template || !encodings) return moves;
  // The transitions' gates count distinct values with a Set, which tells Date objects apart by
  // identity, so time values are compared by their timestamp: a line over 8 days of 5 series
  // has 8 categories for its bar transition, not 40.
  const data = rows.map((row) =>
    Object.fromEntries(
      Object.entries(row).map(([key, value]) => [
        key,
        value instanceof Date ? value.getTime() : value,
      ]),
    ),
  );
  let states: ReturnType<typeof computeChartTypeStates>;
  try {
    states = computeChartTypeStates(
      template,
      encodings,
      data,
      vlGetTemplateDef,
    );
  } catch {
    return moves;
  }
  if (!states) return moves;
  for (const id of states.ids) {
    const chartType = states.chartTypeById[id];
    if (!chartType) continue;
    const moved: WireFlintChartSpec = {
      ...chartSpec,
      chartType,
      encodings: {},
    };
    delete moved.chartProperties;
    for (const [channel, encoding] of Object.entries(states.statesById[id])) {
      if (encoding.field) {
        moved.encodings[channel] = encodingFor(chartSpec, encoding.field);
      }
    }
    moves.set(chartType, moved);
  }
  return moves;
}

/** The spec with a chart property set, or reset to the compiler's default when the choice is undefined. */
export function withProperty(
  chartSpec: WireFlintChartSpec,
  key: string,
  value: unknown,
): WireFlintChartSpec {
  const chartProperties = { ...(chartSpec.chartProperties ?? {}) };
  if (value === undefined) delete chartProperties[key];
  else chartProperties[key] = value;
  return { ...chartSpec, chartProperties };
}

function propertyChoices(option: ChartOption): PanelChoice[] {
  switch (option.type) {
    case "binary":
      return [
        { value: valueKey(true), label: m.mcp_chart_option_on() },
        { value: valueKey(false), label: m.mcp_chart_option_off() },
      ];
    case "discrete":
      return option.options.map((choice) => ({
        value: valueKey(choice.value),
        label: choice.label,
      }));
    default:
      return [];
  }
}

function panelProperties(input: ChartAssemblyInput): PanelProperty[] {
  let options: ChartOption[];
  try {
    options = getChartOptions(input);
  } catch {
    return [];
  }
  return options
    .filter(
      (option) =>
        option.applicable &&
        (option.type === "binary" || option.type === "discrete"),
    )
    .slice(0, MAX_PANEL_PROPERTIES)
    .map((option) => ({
      key: option.key,
      label: option.label,
      choices: propertyChoices(option),
      // Unset is the template's default, which a discrete option lists as a choice, or off.
      value: valueKey(
        option.type === "binary" ? option.value === true : option.value,
      ),
    }));
}

/**
 * Flint's arrange control: the views of the same bindings, stored under its own chart property
 * (`key`). Its moves to another chart type are left out, since the chart-type control covers them.
 */
function panelLayout(input: ChartAssemblyInput): FlintPanel["layout"] {
  let arrange: NonNullable<ReturnType<typeof getChartTransform>>["arrange"];
  try {
    arrange = getChartTransform(input)?.arrange;
  } catch {
    return undefined;
  }
  if (!arrange || arrange.ids.length < 2) return undefined;
  const { key, ids, labels } = arrange;
  const current = ids[arrange.index] ?? ids[0];
  const choices = ids
    .map((id, i) => ({ value: id, label: layoutLabel(id, labels[i] ?? id) }))
    .filter((choice, i) => i < MAX_LAYOUT_CHOICES || choice.value === current);
  return { key, choices, value: current };
}

/**
 * Builds the panel for the chart spec in effect. The chart-type choices are the moves of
 * `baseSpec`, the result's own spec, and a switch always moves from it, so the list stays put
 * while the user tries types and choosing the original type restores the original bindings.
 * While the chart draws a comparison, only the types that can draw it too are offered.
 */
export function buildFlintPanel(
  chartSpec: WireFlintChartSpec,
  compile: FlintCompileInput,
  fields: Record<string, WireField>,
  baseSpec: WireFlintChartSpec = chartSpec,
): FlintPanel {
  const input = assemblyInput(compile);
  const properties = panelProperties(input);
  const layout = panelLayout(input);
  const moves = chartTypeMoves(baseSpec, fields, compile.rows);
  if (compile.comparison) {
    for (const [type, moved] of moves) {
      if (!comparisonChannel(moved, fields)) moves.delete(type);
    }
  }
  const types = [baseSpec.chartType, ...moves.keys()];

  return {
    chartType:
      moves.size > 0
        ? {
            current: chartSpec.chartType,
            choices: types.map((type) => ({ value: type, label: type })),
          }
        : undefined,
    layout,
    properties,
    apply: {
      chartType: (type) => moves.get(type) ?? baseSpec,
      // The first view is the authored one, which needs no override.
      layout: (id) =>
        layout
          ? withProperty(
              chartSpec,
              layout.key,
              id === layout.choices[0].value ? undefined : id,
            )
          : chartSpec,
      // The default choice of a discrete option (`null`) clears the property.
      property: (key, choice) =>
        withProperty(chartSpec, key, JSON.parse(choice) ?? undefined),
    },
  };
}
