import {
  LOADING_CELL,
  SHOW_MORE_BUTTON,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-constants";
import {
  getTimeGrainFromDimension,
  isTimeDimension,
} from "@rilldata/web-common/features/dashboards/pivot/pivot-utils";
import {
  ComparisonModifierSuffixRegex,
  COMPARISON_DELTA,
  COMPARISON_PERCENT,
  COMPARISON_VALUE,
  type PivotDataRow,
  type PivotDataState,
  type PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { ColumnDef, ExpandedState } from "tanstack-table-8-svelte-5";
import { excelNumberFormat, PERCENT_FORMAT } from "./number-format";
import {
  excelDateFormat,
  formatTimeLabel,
  localIsoInZone,
  timeColumnType,
} from "./time-format";
import type { ExportColumn, ExportTable } from "./payload";

interface LeafColumn {
  // The row field the column reads: a measure or dimension name, or a cell
  // key like "c0v1m0" under column dimensions.
  key: string;
  // Headers of the column groups above the leaf, outermost first.
  groups: string[];
  header: string;
  // The measure the column shows, with its comparison suffix if it has one;
  // undefined for dimension columns.
  measure?: string;
}

// Turns the pivot's rendered state into a flat table: the leaf columns of the
// (possibly nested) column definitions, the rows the viewer sees, with expanded
// rows followed by their sub-rows, and the totals row when the pivot shows one.
// Nested row headers share one column, so their time values become labels; a
// flat table's time columns stay typed.
export function flattenPivotData(
  state: PivotDataState,
  config: PivotDataStoreConfig,
): ExportTable {
  const leaves = leafColumns(state.columnDef);
  const timeZone = config.time.timeZone || "UTC";
  const timeDimension = config.time.timeDimension;
  const nestedRowHeader =
    !config.isFlat && config.rowDimensionNames.length > 0
      ? config.rowDimensionNames[0]
      : undefined;

  const visible = visibleRows(state.data, config.pivot.expanded);
  const columns = leaves.map((leaf) =>
    columnFor(
      leaf,
      config,
      nestedRowHeader,
      visible.map(({ row }) => row[leaf.key]),
    ),
  );

  const toCells = (row: PivotDataRow, depth: number) =>
    leaves.map((leaf, i) =>
      cellValue(row[leaf.key], columns[i], leaf.key, depth, {
        config,
        nestedRowHeader,
        timeDimension,
        timeZone,
      }),
    );

  const rows = visible.map(({ row, depth }) => toCells(row, depth));
  const totals = state.totalsRowData;
  if (!totals || config.pivot.showTotalsRow === false) {
    return { columns, rows };
  }
  return {
    columns,
    rows,
    totals: toCells(totalsRow(totals, leaves), 0),
    totalsLast: config.pivot.totalsRowPosition === "bottom",
  };
}

// The leaf columns in display order, each with the headers of the groups above
// it ("US" for a column dimension value). The row-totals group has an empty
// header and is named "Total".
export function leafColumns(
  defs: ColumnDef<PivotDataRow>[],
  groups: string[] = [],
): LeafColumn[] {
  return defs.flatMap((def) => {
    const group = def as ColumnDef<PivotDataRow> & {
      columns?: ColumnDef<PivotDataRow>[];
    };
    const rawHeader = typeof def.header === "string" ? def.header : "";
    const header = rawHeader.replace(/\u00A0/g, "").trim();
    if (group.columns) {
      const label =
        header || (group.columns.length && groups.length === 0 ? "Total" : "");
      return leafColumns(group.columns, label ? [...groups, label] : groups);
    }
    const accessor = (def as { accessorKey?: string }).accessorKey;
    const key = accessor ? String(accessor) : (def.id ?? "");
    // Measure leaves carry the measure's name (see pivot-column-definition.ts),
    // which their key no longer holds under column dimensions.
    const measure = (def as { name?: string }).name;
    return [{ key, groups, header: header || key, measure }];
  });
}

// The rows in the order the table shows them: a row, then its sub-rows when it
// is expanded. Placeholder rows (show more, loading) are skipped.
export function visibleRows(
  rows: PivotDataRow[] | undefined,
  expanded: ExpandedState,
  path: string[] = [],
): { row: PivotDataRow; depth: number }[] {
  if (!rows) return [];
  return rows.flatMap((row, index) => {
    if (isPlaceholderRow(row)) return [];
    const id = [...path, String(index)];
    const isExpanded =
      expanded === true ||
      (typeof expanded === "object" && !!expanded[id.join(".")]);
    const children = isExpanded ? visibleRows(row.subRows, expanded, id) : [];
    return [{ row, depth: path.length }, ...children];
  });
}

// A "show more" or loading row the pivot renders in place of data.
export function isPlaceholderRow(row: PivotDataRow): boolean {
  return Object.values(row).some(
    (value) => value === SHOW_MORE_BUTTON || value === LOADING_CELL,
  );
}

function totalsRow(totals: PivotDataRow, leaves: LeafColumn[]): PivotDataRow {
  const row: PivotDataRow = { ...totals };
  const first = leaves[0]?.key;
  if (
    first &&
    (row[first] === undefined || row[first] === null || row[first] === "")
  ) {
    row[first] = "Total";
  }
  return row;
}

// `values` are the column's values in the visible rows; they size the decimals
// of measures whose format depends on them.
function columnFor(
  leaf: LeafColumn,
  config: PivotDataStoreConfig,
  nestedRowHeader: string | undefined,
  values: unknown[],
): ExportColumn {
  const measureName = leaf.measure ?? leaf.key;
  const base = measureName.replace(ComparisonModifierSuffixRegex, "");
  const measure =
    config.allMeasures.find((m) => m.name === base) ??
    config.ephemeralMeasures?.find((m) => m.name === base);
  const isMeasure = !!measure || config.measureNames.includes(base);
  if (isMeasure) {
    return {
      name: leaf.key,
      displayName: displayName(
        leaf,
        measureHeader(measureName, measure?.displayName || base, leaf.header),
      ),
      type: WorkbookColumnType.NUMBER,
      numberFormat: measureName.endsWith(COMPARISON_PERCENT)
        ? PERCENT_FORMAT
        : excelNumberFormat(measure, values),
    };
  }
  if (
    leaf.key !== nestedRowHeader &&
    isTimeDimension(leaf.key, config.time.timeDimension)
  ) {
    const grain = getTimeGrainFromDimension(leaf.key) as V1TimeGrain;
    return {
      name: leaf.key,
      displayName: displayName(leaf, leaf.header),
      type: timeColumnType(grain),
      numberFormat: excelDateFormat(grain),
    };
  }
  return {
    name: leaf.key,
    displayName: displayName(leaf, leaf.header),
    type: WorkbookColumnType.STRING,
  };
}

// A header with the column's groups folded in: "US / Impressions".
function displayName(leaf: LeafColumn, header: string): string {
  return [...leaf.groups, header].join(" / ");
}

// The header of a measure column. A comparison column's own header ("Δ",
// "Δ %") relies on the measure column beside it, so on a data sheet it names
// the measure, like the leaderboard sheets do.
function measureHeader(
  measureName: string,
  measureLabel: string,
  header: string,
): string {
  if (measureName.endsWith(COMPARISON_DELTA)) return `${measureLabel} change`;
  if (measureName.endsWith(COMPARISON_PERCENT)) {
    return `${measureLabel} % change`;
  }
  if (measureName.endsWith(COMPARISON_VALUE)) {
    return `${measureLabel} (previous)`;
  }
  return header;
}

function cellValue(
  value: unknown,
  column: ExportColumn,
  key: string,
  depth: number,
  ctx: {
    config: PivotDataStoreConfig;
    nestedRowHeader: string | undefined;
    timeDimension: string;
    timeZone: string;
  },
): unknown {
  if (value === undefined || value === null) return null;
  if (Array.isArray(value)) return null;
  if (
    column.type === WorkbookColumnType.DATE ||
    column.type === WorkbookColumnType.TIMESTAMP
  ) {
    return localIsoInZone(value, ctx.timeZone);
  }
  // The row header of a nested pivot holds a different dimension at every depth.
  if (key === ctx.nestedRowHeader && typeof value === "string") {
    const dimension = ctx.config.rowDimensionNames[depth];
    if (isTimeDimension(dimension, ctx.timeDimension)) {
      return formatTimeLabel(
        value,
        getTimeGrainFromDimension(dimension) as V1TimeGrain,
        ctx.timeZone,
      );
    }
  }
  return value;
}
