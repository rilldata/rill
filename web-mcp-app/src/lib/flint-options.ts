import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  assembleVegaLite,
  getChartOptions,
  getChartTransform,
  vlAdaptChart,
  vlGetTemplateDef,
  vlRecommendChartTypes,
} from "flint-chart/vegalite";
import type {
  WireFlintChartSpec,
  WireFlintEncodingValue,
  WireRow,
} from "./contract";
import { stripFlintChartSpec, type FlintCompileInput } from "./flint-input";

/**
 * Presentation controls derived from Flint's own option model, so the panel can never offer a
 * knob the compiler ignores: applicable binary and discrete chart properties (`getChartOptions`),
 * the alternative arrangements of the same bindings (the `arrange` control of
 * `getChartTransform`), and chart types the current bindings can move to
 * (`vlRecommendChartTypes` + `vlAdaptChart`). Every edit is a new chart spec applied
 * client-side; nothing here talks to the server.
 *
 * Adapted from `packages/flint-mcp/ui/src/options.ts` in microsoft/flint-chart (MIT License,
 * Copyright (c) Microsoft Corporation), restricted to the controls this widget renders.
 */

type AssemblyInput = Parameters<typeof assembleVegaLite>[0];

/** The shape of a Flint chart option this module reads; declared locally, like `compile.ts` does. */
interface RawOption {
  key: string;
  label: string;
  type: "continuous" | "discrete" | "binary";
  applicable: boolean;
  value?: unknown;
  options?: { value: unknown; label: string }[];
}

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
  /** Chart types the data can move to; undefined when Flint recommends none that compile. */
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

/** The toolbar stays readable: a few properties, a few views, a few chart types. */
export const MAX_PANEL_PROPERTIES = 3;
export const MAX_LAYOUT_CHOICES = 4;
export const MAX_CHART_TYPE_CANDIDATES = 4;
/** Flint types the server does not plan queries for in this phase. */
export const EXCLUDED_CHART_TYPES: ReadonlySet<string> = new Set([
  "Map",
  "Choropleth",
  "KPI Card",
]);

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

function assemblyInput(
  compile: FlintCompileInput,
  rows: WireRow[],
): AssemblyInput {
  return {
    data: { values: rows },
    semantic_types: compile.fields.semantic_types,
    field_display_names: compile.fields.field_display_names,
    chart_spec: compile.spec,
    options: { addTooltips: true },
  } as AssemblyInput;
}

/** The single-field bindings of a spec, as `vlAdaptChart` takes them; series folds are skipped. */
function singleBindings(chartSpec: WireFlintChartSpec): Record<string, string> {
  const bindings: Record<string, string> = {};
  for (const [channel, value] of Object.entries(chartSpec.encodings ?? {})) {
    const field =
      typeof value === "string"
        ? value
        : Array.isArray(value)
          ? undefined
          : value?.field;
    if (field) bindings[channel] = field;
  }
  return bindings;
}

/**
 * The binding a moved field keeps: its Rill `limit`, so the top-N survives the move. Its sort
 * keys are dropped, since `sortBy` names a channel of the old chart type and the new type's
 * own ordering applies.
 */
function encodingFor(
  chartSpec: WireFlintChartSpec,
  field: string,
): WireFlintEncodingValue {
  for (const value of Object.values(chartSpec.encodings ?? {})) {
    if (
      typeof value === "object" &&
      !Array.isArray(value) &&
      value.field === field
    ) {
      return value.limit === undefined ? field : { field, limit: value.limit };
    }
  }
  return field;
}

/** The spec moved to another chart type, with Flint's re-binding of the fields. */
export function withChartType(
  chartSpec: WireFlintChartSpec,
  chartType: string,
  rows: WireRow[],
  semanticTypes: Record<string, string>,
): WireFlintChartSpec {
  const adapted = vlAdaptChart(
    chartSpec.chartType,
    chartType,
    singleBindings(chartSpec),
    rows,
    semanticTypes,
  );
  const encodings: Record<string, WireFlintEncodingValue> = {};
  for (const [channel, field] of Object.entries(adapted)) {
    if (field) encodings[channel] = encodingFor(chartSpec, field);
  }
  // Properties are template-specific; the new type starts from its own defaults.
  const moved: WireFlintChartSpec = { ...chartSpec, chartType, encodings };
  delete moved.chartProperties;
  return moved;
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

function propertyChoices(option: RawOption): PanelChoice[] {
  if (option.type === "binary") {
    return [
      { value: valueKey(true), label: m.mcp_chart_option_on() },
      { value: valueKey(false), label: m.mcp_chart_option_off() },
    ];
  }
  return (option.options ?? []).map((choice) => ({
    value: valueKey(choice.value),
    label: choice.label,
  }));
}

function panelProperties(input: AssemblyInput): PanelProperty[] {
  let options: RawOption[];
  try {
    options = getChartOptions(input) as unknown as RawOption[];
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
function panelLayout(input: AssemblyInput): FlintPanel["layout"] {
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
 * Chart types the current bindings can move to: Flint's ranking for the data, kept when the
 * adapted spec compiles, since a recommendation is made for the dataset rather than for these
 * bindings.
 */
function chartTypeCandidates(
  chartSpec: WireFlintChartSpec,
  compile: FlintCompileInput,
  rows: WireRow[],
): string[] {
  let ranked: string[];
  try {
    ranked = vlRecommendChartTypes(rows, compile.fields.semantic_types, {
      max: 12,
    });
  } catch {
    return [];
  }
  const candidates: string[] = [];
  for (const type of ranked) {
    if (candidates.length >= MAX_CHART_TYPE_CANDIDATES) break;
    if (type === chartSpec.chartType || EXCLUDED_CHART_TYPES.has(type))
      continue;
    if (!vlGetTemplateDef(type)) continue;
    try {
      const moved = withChartType(
        chartSpec,
        type,
        rows,
        compile.fields.semantic_types,
      );
      if (Object.keys(moved.encodings).length === 0) continue;
      assembleVegaLite(
        assemblyInput({ ...compile, spec: stripFlintChartSpec(moved) }, rows),
      );
      candidates.push(type);
    } catch {
      // Flint cannot draw the data as this type; skip it.
    }
  }
  return candidates;
}

/**
 * Builds the panel for the chart spec in effect. The chart-type choices are computed from
 * `baseSpec`, the result's own spec, and a switch always moves from it, so the list stays put
 * while the user tries types and choosing the original type restores the original bindings.
 */
export function buildFlintPanel(
  chartSpec: WireFlintChartSpec,
  compile: FlintCompileInput,
  rows: WireRow[],
  baseSpec: WireFlintChartSpec = chartSpec,
): FlintPanel {
  const input = assemblyInput(compile, rows);
  const properties = panelProperties(input);
  const layout = panelLayout(input);
  const candidates = chartTypeCandidates(baseSpec, compile, rows);
  const types = [baseSpec.chartType, ...candidates];
  if (!types.includes(chartSpec.chartType)) types.push(chartSpec.chartType);

  return {
    chartType:
      candidates.length > 0
        ? {
            current: chartSpec.chartType,
            choices: types.map((type) => ({ value: type, label: type })),
          }
        : undefined,
    layout,
    properties,
    apply: {
      chartType: (type) =>
        type === baseSpec.chartType
          ? baseSpec
          : withChartType(baseSpec, type, rows, compile.fields.semantic_types),
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
