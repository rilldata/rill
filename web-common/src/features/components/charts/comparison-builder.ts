import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type {
  Field,
  NumericMarkPropDef,
  OffsetDef,
} from "vega-lite/types_unstable/channeldef.js";
import type { Transform } from "vega-lite/types_unstable/transform.js";

export const MeasureKeyField = "measure_key";
export const ColorWithComparisonField = "color_with_comparison";
export const SortOrderField = "sortOrder";

/**
 * Creates Vega-Lite transforms for period-over-period comparison.
 * This function generates the fold and calculate transforms needed to
 * display current and comparison data side-by-side.
 *
 * The data structure has current and previous values as separate measure fields:
 * e.g., "total_bids" (current) and "total_bids_prev" (comparison).
 *
 * Field names are passed raw (not pre-escaped) on purpose. `fold.as` stores
 * its second element as a literal field name, while downstream consumers
 * such as `pivot.value` resolve field-path escapes — pre-escaping here makes
 * those two views of the same field disagree (e.g. `"x \\| y"` vs `"x | y"`).
 * Embedded Vega expression literals escape `\` and `'` via `escapeVegaString`.
 */
export function createComparisonTransforms(
  xField: string | undefined,
  measureField: string | undefined,
  colorField?: string,
): Transform[] {
  if (!xField || !measureField) return [];

  const previousMeasureField = measureField + ComparisonDeltaPreviousSuffix;
  const previousLiteral = escapeVegaString(previousMeasureField);
  const transforms: Transform[] = [];

  // Fold the current and previous measure fields into the original measure
  // column, with a `measure_key` flag indicating which period each row is.
  transforms.push({
    fold: [measureField, previousMeasureField],
    as: [MeasureKeyField, measureField],
  });

  if (colorField) {
    const colorLiteral = escapeVegaString(colorField);
    // Synthetic nominal field for grouping current and previous together.
    transforms.push({
      calculate: `datum['${colorLiteral}'] + (datum['${MeasureKeyField}'] === '${previousLiteral}' ? '${ComparisonDeltaPreviousSuffix}' : '')`,
      as: ColorWithComparisonField,
    });

    // Sort order groups by color first, then by period (current before previous).
    transforms.push({
      calculate: `datum['${colorLiteral}'] + '_' + (datum['${MeasureKeyField}'] === '${previousLiteral}' ? '1' : '0')`,
      as: SortOrderField,
    });
  } else {
    transforms.push({
      calculate: `datum['${MeasureKeyField}'] === '${previousLiteral}' ? 1 : 0`,
      as: SortOrderField,
    });
  }

  return transforms;
}

/**
 * Creates an opacity encoding for comparison mode.
 *
 * `measureField` must be the raw measure name; the comparison-period key
 * compared against here is derived the same way as in
 * `createComparisonTransforms` so the test stays in sync.
 */
export function createComparisonOpacityEncoding(
  measureField: string | undefined,
): NumericMarkPropDef<Field> {
  if (!measureField) return { value: 1 };

  const previousLiteral = escapeVegaString(
    measureField + ComparisonDeltaPreviousSuffix,
  );
  return {
    condition: [
      {
        test: `datum['${MeasureKeyField}'] === '${previousLiteral}'`,
        value: 0.4,
      },
    ],
    value: 1,
  };
}

// Escape `\` and `'` for safe embedding inside a Vega expression
// single-quoted string literal.
function escapeVegaString(value: string): string {
  return value.replace(/\\/g, "\\\\").replace(/'/g, "\\'");
}

/**
 * Creates an xOffset encoding for comparison mode.
 * This positions current and comparison bars/lines side-by-side.
 */
export function createComparisonXOffsetEncoding(): OffsetDef<Field> {
  return {
    field: MeasureKeyField,
    sort: { field: SortOrderField },
  };
}
