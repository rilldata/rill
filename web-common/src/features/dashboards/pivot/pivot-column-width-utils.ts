import { extractSamples } from "@rilldata/web-common/components/virtualized-table/init-widths";
import { isTimeDimension } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import type { PivotDataRow } from "@rilldata/web-common/features/dashboards/pivot/types";
import { clamp } from "@rilldata/web-common/lib/clamp";

export const MIN_COL_WIDTH = 150;
export const MAX_COL_WIDTH = 600;
export const MAX_INIT_COL_WIDTH = 400;

export const MIN_MEASURE_WIDTH = 70;
export const MAX_MEASURE_WIDTH = 300;
export const INIT_MEASURE_WIDTH = 100;
export const MEASURE_PADDING = 24;

export function calculateFirstColumnWidth(
  firstColumnName: string,
  timeDimension: string,
  dataRows: PivotDataRow[],
) {
  // Dates are displayed as shorter values
  if (isTimeDimension(firstColumnName, timeDimension)) return MIN_COL_WIDTH;

  const samples = extractSamples(
    dataRows.map((row) => row[firstColumnName]),
  ).filter((v): v is string => typeof v === "string");

  const maxValueLength = samples.reduce((max, value) => {
    return Math.max(max, value.length);
  }, 0);

  const finalBasis = Math.max(firstColumnName.length, maxValueLength);
  const pixelLength = finalBasis * 8;
  const final = clamp(MIN_COL_WIDTH, pixelLength + 16, MAX_INIT_COL_WIDTH);

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
) {
  let maxValueLength: number;
  if (totalsRow) {
    const value = totalsRow[measureName];
    if (typeof value === "string" || typeof value === "number") {
      maxValueLength = String(formatter(value)).length;
    } else {
      maxValueLength = 0;
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

  const finalBasis = Math.max(label.length, maxValueLength);
  const pixelLength = finalBasis * 7;
  return clamp(
    MIN_MEASURE_WIDTH,
    pixelLength + MEASURE_PADDING,
    MAX_MEASURE_WIDTH,
  );
}
