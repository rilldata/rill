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

export type ColumnWidthRole = "dimension" | "measure";

/**
 * The resizer bounds for a column role. Configured widths are clamped into
 * the same bounds, so a later drag never snaps a configured width back.
 */
export function roleWidthBounds(role: ColumnWidthRole): {
  min: number;
  max: number;
} {
  return role === "measure"
    ? {
        min: COLUMN_WIDTH_CONSTANTS.MIN_MEASURE_WIDTH,
        max: COLUMN_WIDTH_CONSTANTS.MAX_MEASURE_WIDTH,
      }
    : {
        min: COLUMN_WIDTH_CONSTANTS.MIN_COL_WIDTH,
        max: COLUMN_WIDTH_CONSTANTS.MAX_COL_WIDTH,
      };
}

export function clampColumnWidth(role: ColumnWidthRole, width: number) {
  const { min, max } = roleWidthBounds(role);
  return clamp(min, Math.round(width), max);
}

export type FillWidthColumn = {
  width: number;
  role: ColumnWidthRole;
  // Pinned columns (configured or resized by the user) take no extra width.
  pinned?: boolean;
};

const DIMENSION_FILL_WEIGHT = 1.4;
const MEASURE_FILL_WEIGHT = 1;

export function distributeColumnWidthsToFillContainer(
  columns: FillWidthColumn[],
  containerWidth: number,
) {
  const baseWidths = columns.map(({ width }) => width);
  const totalWidth = baseWidths.reduce((sum, width) => sum + width, 0);
  const extraWidth = containerWidth - totalWidth;

  if (!columns.length || extraWidth <= 0) return baseWidths;

  const flexible = columns.filter(({ pinned }) => !pinned);
  const hasDimensions = flexible.some(({ role }) => role === "dimension");
  const hasMeasures = flexible.some(({ role }) => role === "measure");
  const weights = columns.map(({ role, pinned }) => {
    if (pinned) return 0;
    return hasDimensions && hasMeasures && role === "dimension"
      ? DIMENSION_FILL_WEIGHT
      : MEASURE_FILL_WEIGHT;
  });
  const totalWeight = weights.reduce((sum, weight) => sum + weight, 0);
  if (totalWeight === 0) return baseWidths;

  return baseWidths.map(
    (width, i) => width + (extraWidth * weights[i]) / totalWeight,
  );
}

export type FitWidthColumn = {
  width: number;
  min: number;
  max: number;
};

/**
 * One-shot resize of every column so the table fits the available width.
 * When the columns overflow, each shrinks in proportion to its slack above
 * its minimum; when they underflow, each grows in proportion to its headroom
 * below its maximum. Bounds are never crossed, so if even the minimums
 * overflow the table still scrolls by the remainder. Results are whole pixels
 * and never sum to more than the available width when a fit is feasible.
 */
export function fitColumnWidthsToContainer(
  columns: FitWidthColumn[],
  availableWidth: number,
): number[] {
  if (!columns.length) return [];

  const totalWidth = columns.reduce((sum, { width }) => sum + width, 0);
  const delta = availableWidth - totalWidth;
  if (delta === 0) return columns.map(({ width }) => width);

  // Positive delta grows toward max; negative delta shrinks toward min.
  const room = columns.map(({ width, min, max }) =>
    delta > 0 ? Math.max(0, max - width) : Math.max(0, width - min),
  );
  const totalRoom = room.reduce((sum, r) => sum + r, 0);
  if (totalRoom === 0) return columns.map(({ width }) => width);

  // Never move further than the room allows, so bounds are respected.
  const applied = Math.sign(delta) * Math.min(Math.abs(delta), totalRoom);
  return columns.map(({ width }, i) =>
    Math.floor(width + (applied * room[i]) / totalRoom),
  );
}

export type LayoutColumn = FillWidthColumn;

/**
 * Display widths for a table. When the base widths overflow the container and
 * `fit` is on, every unpinned column shrinks proportionally within its role
 * bounds while pinned columns keep their width; the table still scrolls if the
 * pinned widths and minimums alone overflow. Otherwise, when `fill` is on, the
 * extra width is spread over the unpinned columns. Otherwise the base widths
 * are used as they are.
 */
export function layoutColumnWidths(
  columns: LayoutColumn[],
  containerWidth: number,
  options: { fill: boolean; fit: boolean },
): number[] {
  if (!columns.length) return [];

  const totalWidth = columns.reduce((sum, { width }) => sum + width, 0);
  if (options.fit && containerWidth > 0 && totalWidth > containerWidth) {
    return fitColumnWidthsToContainer(
      columns.map(({ width, role, pinned }) =>
        pinned
          ? { width, min: width, max: width }
          : { width, ...roleWidthBounds(role) },
      ),
      containerWidth,
    );
  }
  if (options.fill) {
    return distributeColumnWidthsToFillContainer(columns, containerWidth);
  }
  return columns.map(({ width }) => width);
}

/**
 * Width estimate for a dimension column. Values are sampled by the column id
 * (the field name the data rows are keyed by), while the header basis uses the
 * display label, which may differ from the id.
 */
export function calculateColumnWidth(
  columnId: string,
  label: string,
  timeDimension: string,
  dataRows: PivotDataRow[],
) {
  // Dates are displayed as shorter values
  if (isTimeDimension(columnId, timeDimension))
    return COLUMN_WIDTH_CONSTANTS.MIN_COL_WIDTH;

  const samples = extractSamples(dataRows.map((row) => row[columnId])).filter(
    (v): v is string => typeof v === "string",
  );

  const maxValueLength = samples.reduce((max, value) => {
    return Math.max(max, value.length);
  }, 0);

  const finalBasis = Math.max(label.length, maxValueLength);
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

/**
 * Width estimate for the nested table's row header column. The values are
 * sampled by the first row dimension's name; the label is the joined label of
 * all row dimensions.
 */
export function calculateRowDimensionWidth(
  columnId: string,
  label: string,
  timeDimension: string,
  dataRows: PivotDataRow[],
) {
  let width = COLUMN_WIDTH_CONSTANTS.ROW_DIMENSION_MIN_WIDTH;
  if (!isTimeDimension(columnId, timeDimension)) {
    width = calculateColumnWidth(columnId, label, timeDimension, dataRows);
  }

  return clamp(
    COLUMN_WIDTH_CONSTANTS.ROW_DIMENSION_MIN_WIDTH,
    width,
    COLUMN_WIDTH_CONSTANTS.MAX_INIT_COL_WIDTH,
  );
}

export function getNestedRowDimensionWidthKey(
  widthScopeKey: string,
  rowDimensions: Array<{ name: string }>,
) {
  const rowDimensionName = rowDimensions[0]?.name;
  if (!widthScopeKey || !rowDimensionName) return undefined;

  return `${widthScopeKey}:${rowDimensionName}`;
}
