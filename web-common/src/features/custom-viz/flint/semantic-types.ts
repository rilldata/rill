import {
  MetricsViewSpecDimensionType,
  V1TimeGrain,
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { includesCurrencySymbol } from "@rilldata/web-common/lib/number-formatting/utils/d3-format-utils";

/**
 * Derives Flint's per-field metadata from a metrics view.
 *
 * Flint asks callers to declare a semantic type per field, and uses it to pick aggregation defaults,
 * scale type, stacking behaviour, number formats, sort order and diverging midpoints. Rill already
 * knows all of this from the metrics view, so components never declare it: the author writes only a
 * chart type and field bindings, and the semantics are filled in per binding. That is why the same
 * component renders correctly against different metrics views.
 *
 * Only the strings in Flint's own registry do anything — anything else silently degrades to
 * nominal/plain — so `semantic-types.spec.ts` asserts every value this module can emit is registered.
 */

/** Semantic types this module can emit. Must all exist in Flint's registry. */
export type FlintSemanticType =
  | "Year"
  | "YearQuarter"
  | "YearMonth"
  | "YearWeek"
  | "Date"
  | "DateTime"
  | "Amount"
  | "Percentage"
  | "Duration"
  | "Count"
  | "Number"
  | "Category";

export interface FlintFields {
  semantic_types: Record<string, FlintSemanticType>;
  field_display_names: Record<string, string>;
}

/**
 * Maps the grain a time dimension is bucketed at to the matching Flint temporal type.
 * Grains finer than a day have no dedicated Flint type and fall back to DateTime.
 */
const TIME_GRAIN_TO_SEMANTIC_TYPE: Partial<
  Record<V1TimeGrain, FlintSemanticType>
> = {
  [V1TimeGrain.TIME_GRAIN_YEAR]: "Year",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "YearQuarter",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "YearMonth",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "YearWeek",
  [V1TimeGrain.TIME_GRAIN_DAY]: "Date",
};

export function deriveFlintFields(
  columns: string[],
  metricsViewSpec: V1MetricsViewSpec | undefined,
  timeGrain?: V1TimeGrain,
): FlintFields {
  const semantic_types: Record<string, FlintSemanticType> = {};
  const field_display_names: Record<string, string> = {};

  if (!metricsViewSpec) return { semantic_types, field_display_names };

  const measures = new Map(
    (metricsViewSpec.measures ?? []).map((m) => [m.name as string, m]),
  );
  const dimensions = new Map(
    (metricsViewSpec.dimensions ?? []).map((d) => [d.name as string, d]),
  );

  for (const column of columns) {
    const measure = measures.get(column);
    if (measure) {
      semantic_types[column] = measureSemanticType(measure);
      if (measure.displayName)
        field_display_names[column] = measure.displayName;
      continue;
    }

    const dimension = dimensions.get(column);
    if (dimension) {
      semantic_types[column] = dimensionSemanticType(
        column,
        dimension,
        metricsViewSpec,
        timeGrain,
      );
      if (dimension.displayName) {
        field_display_names[column] = dimension.displayName;
      }
      continue;
    }

    // A column that is neither a declared measure nor dimension: most often the metrics view's
    // primary time dimension, which is not always present in the dimensions list.
    if (column === metricsViewSpec.timeDimension) {
      semantic_types[column] = timeSemanticType(timeGrain);
    }
    // Otherwise leave it unannotated so Flint infers from the raw values.
  }

  return { semantic_types, field_display_names };
}

/**
 * Picks a Flint type from how the measure is formatted rather than from its runtime data type:
 * the format preset is what encodes intent (money, ratio, elapsed time), and it is also what the
 * frontend formatters key off.
 */
function measureSemanticType(
  measure: MetricsViewSpecMeasure,
): FlintSemanticType {
  switch (measure.formatPreset as FormatPreset | undefined) {
    case FormatPreset.CURRENCY_USD:
    case FormatPreset.CURRENCY_EUR:
      return "Amount";
    case FormatPreset.PERCENTAGE:
      return "Percentage";
    case FormatPreset.INTERVAL:
      return "Duration";
  }

  if (measure.formatD3) {
    if (includesCurrencySymbol(measure.formatD3)) return "Amount";
    if (measure.formatD3.includes("%")) return "Percentage";
  }

  // Integer measures are counts often enough that the integer tick strategy is the better default.
  if (measure.dataType?.code && INTEGERS.has(measure.dataType.code)) {
    return "Count";
  }

  return "Number";
}

function dimensionSemanticType(
  column: string,
  dimension: MetricsViewSpecDimension,
  metricsViewSpec: V1MetricsViewSpec,
  timeGrain: V1TimeGrain | undefined,
): FlintSemanticType {
  const isTime =
    dimension.type === MetricsViewSpecDimensionType.DIMENSION_TYPE_TIME ||
    column === metricsViewSpec.timeDimension;

  return isTime ? timeSemanticType(timeGrain) : "Category";
}

function timeSemanticType(
  timeGrain: V1TimeGrain | undefined,
): FlintSemanticType {
  if (!timeGrain) return "DateTime";
  return TIME_GRAIN_TO_SEMANTIC_TYPE[timeGrain] ?? "DateTime";
}
