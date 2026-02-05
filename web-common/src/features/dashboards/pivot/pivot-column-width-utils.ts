import { extractSamples } from "@rilldata/web-common/components/virtualized-table/init-widths";
import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  COMPARISON_PERCENT,
  type PivotDataRow,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { clamp } from "@rilldata/web-common/lib/clamp";

export const COLUMN_WIDTH_CONSTANTS = {
  MIN_COL_WIDTH: 100,
  MAX_COL_WIDTH: 600,
  MAX_INIT_COL_WIDTH: 400,
  MIN_MEASURE_WIDTH: 60,
  MAX_MEASURE_WIDTH: 300,
  INIT_MEASURE_WIDTH: 100,
  MEASURE_PADDING: 24,
  ROW_DIMENSION_MIN_WIDTH: 160,
  MAX_COL_DIMENSION_HEADER_LENGTH: 18,
};

export function calculateColumnWidth(
  columnName: string,
  timeDimension: string,
  dataRows: PivotDataRow[],
) {
  // Dates are displayed as shorter values
  if (isTimeDimension(columnName, timeDimension))
    return COLUMN_WIDTH_CONSTANTS.MIN_COL_WIDTH;

  const samples = extractSamples(dataRows.map((row) => row[columnName])).filter(
    (v): v is string => typeof v === "string",
  );

  const maxValueLength = samples.reduce((max, value) => {
    return Math.max(max, value.length);
  }, 0);

  const finalBasis = Math.max(columnName.length, maxValueLength);
  const pixelLength = finalBasis * 7;
  const final = clamp(
    COLUMN_WIDTH_CONSTANTS.MIN_COL_WIDTH,
    pixelLength + 16,
    COLUMN_WIDTH_CONSTANTS.MAX_INIT_COL_WIDTH,
  );

  return final;
}

/**
 * For measure column if available, use the totals row data as the heuristic
 * for determining the column width. In most cases the totals row
 * will have the max or close. In absence of the totals row use the
 * data rows for getting a sample
 */
export function calculateMeasureWidth(
  measureName: string,
  label: string,
  formatter: (
    value: string | number | null | undefined,
  ) => string | (null | undefined),
  totalsRow: PivotDataRow | undefined,
  dataRows: PivotDataRow[],
  columnDimensionHeader?: string,
) {
  let maxValueLength: number;
  if (totalsRow) {
    const isPercent = measureName.endsWith(COMPARISON_PERCENT);
    if (isPercent) {
      maxValueLength = 5;
    } else {
      const value = totalsRow[measureName];
      if (typeof value === "string" || typeof value === "number") {
        maxValueLength = String(formatter(value)).length;
      } else {
        maxValueLength = 8;
      }
    }
  } else {
    const samples = extractSamples(
      dataRows.map((row) => row[measureName]),
    ).filter(
      (v): v is string | number =>
        typeof v === "string" || typeof v === "number",
    );

    maxValueLength = samples.reduce((max: number, value) => {
      const stringLength = String(formatter(value)).length;
      return Math.max(max, stringLength);
    }, 0) as number;
  }

  // When there's a column dimension, also consider its header length
  const columnDimensionLength = Math.min(
    columnDimensionHeader?.length ?? 0,
    COLUMN_WIDTH_CONSTANTS.MAX_COL_DIMENSION_HEADER_LENGTH,
  );

  const finalBasis = Math.max(
    label.length,
    maxValueLength,
    columnDimensionLength,
  );
  const pixelLength = finalBasis * 7;
  return clamp(
    COLUMN_WIDTH_CONSTANTS.MIN_MEASURE_WIDTH,
    pixelLength + COLUMN_WIDTH_CONSTANTS.MEASURE_PADDING,
    COLUMN_WIDTH_CONSTANTS.MAX_MEASURE_WIDTH,
  );
}

export function calculateRowDimensionWidth(
  columnName: string,
  timeDimension: string,
  dataRows: PivotDataRow[],
) {
  let width = COLUMN_WIDTH_CONSTANTS.ROW_DIMENSION_MIN_WIDTH;
  if (!isTimeDimension(columnName, timeDimension)) {
    width = calculateColumnWidth(columnName, timeDimension, dataRows);
  }

  return clamp(
    COLUMN_WIDTH_CONSTANTS.ROW_DIMENSION_MIN_WIDTH,
    width,
    COLUMN_WIDTH_CONSTANTS.MAX_INIT_COL_WIDTH,
  );
}
