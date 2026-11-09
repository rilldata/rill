import type { CanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import type {
  ChartDataResult,
  ChartType,
} from "@rilldata/web-common/features/components/charts/types";
import type { LeaderboardItemData } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import type {
  PivotDataState,
  PivotDataStoreConfig,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import type { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1MetricsViewAggregationResponseDataItem,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

// What a display component knows about the data it renders. Each display
// component publishes this on its canvas component instance (see
// BaseCanvasComponent.exportPayload) so the workbook export can read the
// numbers behind every block without re-running queries.
export type ComponentExportPayload =
  | {
      kind: "chart";
      chartType: ChartType;
      spec: CanvasChartSpec;
      data: ChartDataResult;
    }
  | { kind: "kpi"; values: KpiExportValue[] }
  | { kind: "table"; state: PivotDataState; config: PivotDataStoreConfig }
  | { kind: "leaderboard"; tables: LeaderboardExportTable[] }
  | { kind: "customChart"; results: CustomChartExportResult[] }
  | {
      kind: "map";
      rows: V1MetricsViewAggregationResponseDataItem[];
      dimensions: MetricsViewSpecDimension[];
      measures: MetricsViewSpecMeasure[];
    }
  | { kind: "text"; text: string }
  | { kind: "imageUrl"; url: string }
  | { kind: "error"; message: string };

export interface KpiExportValue {
  measure: MetricsViewSpecMeasure | undefined;
  label: string;
  value: number | null;
  comparisonLabel: string | undefined;
  comparisonValue: number | null;
  delta: number | null;
  deltaPercent: number | null;
  // The sparkline's time series (ISO timestamps as the runtime returns them), when the KPI shows one.
  trend?: KpiTrendPoint[];
  timeGrain?: V1TimeGrain;
}

export interface KpiTrendPoint {
  time: string;
  value: number | null;
}

export interface LeaderboardExportTable {
  dimension: MetricsViewSpecDimension;
  measures: MetricsViewSpecMeasure[];
  rows: LeaderboardItemData[];
}

export interface CustomChartExportResult {
  data: Record<string, unknown>[] | undefined;
  schema: VirtualizedTableColumns[] | undefined;
}

// A typed table bound for a data sheet; see collect.ts for how it becomes a WorkbookDataSheet.
export interface ExportColumn {
  name: string;
  displayName: string;
  type: WorkbookColumnType;
  numberFormat?: string;
}

export interface ExportTable {
  columns: ExportColumn[];
  rows: unknown[][];
  // The totals row, kept apart from the rows so the data sheet can write it
  // below its Excel table rather than in it.
  totals?: unknown[];
  // Number formats by row, for a table whose rows hold different measures
  // (the KPIs sheet); see WorkbookDataSheet.row_number_formats.
  rowNumberFormats?: string[];
  // True when the source has more rows than the table holds (the component
  // had not loaded them, or the export's row cap trimmed them).
  moreRows?: boolean;
}
