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
  COMPARISON_PERCENT,
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
  key: string;
  header: string;
}

// Turns the pivot's rendered state into a flat table: the leaf columns of the
// (possibly nested) column definitions, and the rows the viewer sees, with
// expanded rows followed by their sub-rows and the totals row where the pivot
// shows it. Nested row headers share one column, so their time values become
// labels; a flat table's time columns stay typed.
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

  const columns = leaves.map((leaf) =>
    columnFor(leaf, config, nestedRowHeader),
  );

  const rows: unknown[][] = [];
  const pushRow = (row: PivotDataRow, depth: number) => {
    rows.push(
      leaves.map((leaf, i) =>
        cellValue(row[leaf.key], columns[i], leaf.key, depth, {
          config,
          nestedRowHeader,
          timeDimension,
          timeZone,
        }),
      ),
    );
  };

  const totals = state.totalsRowData;
  const showTotals = !!totals && config.pivot.showTotalsRow !== false;
  if (showTotals && config.pivot.totalsRowPosition !== "bottom") {
    pushRow(totalsRow(totals, leaves), 0);
  }
  for (const { row, depth } of visibleRows(state.data, config.pivot.expanded)) {
    pushRow(row, depth);
  }
  if (showTotals && config.pivot.totalsRowPosition === "bottom") {
    pushRow(totalsRow(totals, leaves), 0);
  }

  return { columns, rows };
}

// The leaf columns in display order, with group headers folded into the leaf
// header ("US / Impressions"). The row-totals group has an empty header.
export function leafColumns(
  defs: ColumnDef<PivotDataRow>[],
  prefix: string[] = [],
): LeafColumn[] {
  return defs.flatMap((def) => {
    const group = def as ColumnDef<PivotDataRow> & {
      columns?: ColumnDef<PivotDataRow>[];
    };
    const rawHeader = typeof def.header === "string" ? def.header : "";
    const header = rawHeader.replace(/\u00A0/g, "").trim();
    if (group.columns) {
      const label =
        header || (group.columns.length && prefix.length === 0 ? "Total" : "");
      return leafColumns(group.columns, label ? [...prefix, label] : prefix);
    }
    const accessor = (def as { accessorKey?: string }).accessorKey;
    const key = accessor ? String(accessor) : (def.id ?? "");
    return [{ key, header: [...prefix, header || key].join(" / ") }];
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

function columnFor(
  leaf: LeafColumn,
  config: PivotDataStoreConfig,
  nestedRowHeader: string | undefined,
): ExportColumn {
  const base = leaf.key.replace(ComparisonModifierSuffixRegex, "");
  const measure =
    config.allMeasures.find((m) => m.name === base) ??
    config.ephemeralMeasures?.find((m) => m.name === base);
  const isMeasure = !!measure || config.measureNames.includes(base);
  if (isMeasure) {
    return {
      name: leaf.key,
      displayName: leaf.header,
      type: WorkbookColumnType.NUMBER,
      numberFormat: leaf.key.endsWith(COMPARISON_PERCENT)
        ? PERCENT_FORMAT
        : excelNumberFormat(measure),
    };
  }
  if (
    leaf.key !== nestedRowHeader &&
    isTimeDimension(leaf.key, config.time.timeDimension)
  ) {
    const grain = getTimeGrainFromDimension(leaf.key) as V1TimeGrain;
    return {
      name: leaf.key,
      displayName: leaf.header,
      type: timeColumnType(grain),
      numberFormat: excelDateFormat(grain),
    };
  }
  return {
    name: leaf.key,
    displayName: leaf.header,
    type: WorkbookColumnType.STRING,
  };
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
