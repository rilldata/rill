import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type {
  Field,
  NumericMarkPropDef,
  OffsetDef,
} from "vega-lite/build/src/channeldef";
import type { Transform } from "vega-lite/build/src/transform";

export const MeasureKeyField = "measure_key";
export const ColorWithComparisonField = "color_with_comparison";
export const SortOrderField = "sortOrder";

/**
 * Creates Vega-Lite transforms for period-over-period comparison.
 * This function generates the fold and calculate transforms needed to
 * display current and comparison data side-by-side.
 *
 * The data structure has current and previous values as separate measure fields:
 * e.g., "total_bids" (current) and "total_bids_prev" (comparison)
 */
export function createComparisonTransforms(
  xField: string | undefined,
  measureField: string | undefined,
  colorField?: string,
): Transform[] {
  if (!xField || !measureField) return [];

  const sanitizedMeasure = sanitizeValueForVega(measureField);
  const sanitizedComparisonMeasure = sanitizeValueForVega(
    measureField + ComparisonDeltaPreviousSuffix,
  );
  const transforms: Transform[] = [];

  // Fold the current and previous measure fields
  // This creates two rows per original row: one for current, one for previous
  transforms.push({
    fold: [sanitizedMeasure, sanitizedComparisonMeasure],
    as: [MeasureKeyField, sanitizedMeasure],
  });

  // If there's a color field, create a synthetic nominal field for grouping
  // This combines the color dimension with the comparison key for proper stacking
  if (colorField) {
    const sanitizedColor = sanitizeValueForVega(colorField);
    transforms.push({
      calculate: `datum['${sanitizedColor}'] + (datum['${MeasureKeyField}'] === '${sanitizedComparisonMeasure}' ? '${ComparisonDeltaPreviousSuffix}' : '')`,
      as: ColorWithComparisonField,
    });

    // Add a sort order field that groups by color first, then by period
    // This ensures the order is: A_current, A_previous, B_current, B_previous
    transforms.push({
      calculate: `datum['${sanitizedColor}'] + '_' + (datum['${MeasureKeyField}']=== '${sanitizedComparisonMeasure}' ? '1' : '0')`,
      as: SortOrderField,
    });
  } else {
    // Add a sort order field to ensure current appears before comparison
    transforms.push({
      calculate: `datum['${MeasureKeyField}'] === '${sanitizedComparisonMeasure}' ? 1 : 0`,
      as: SortOrderField,
    });
  }

  return transforms;
}

/**
 * Creates an opacity encoding for comparison mode.
 */
export function createComparisonOpacityEncoding(
  measureField: string,
): NumericMarkPropDef<Field> {
  const sanitizedComparisonMeasure = sanitizeValueForVega(
    measureField + ComparisonDeltaPreviousSuffix,
  );
  return {
    condition: [
      {
        test: `datum['${MeasureKeyField}'] === '${sanitizedComparisonMeasure}'`,
        value: 0.4,
      },
    ],
    value: 1,
  };
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
