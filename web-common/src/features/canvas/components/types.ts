import type { CartesianCanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts/variants/CartesianChart";
import type { CircularCanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts/variants/CircularChart";
import type { KPIGridSpec } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type {
  ComboChartSpec,
  FunnelChartSpec,
  HeatmapChartSpec,
} from "@rilldata/web-common/features/components/charts";
import type { ChartType } from "../../components/charts/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { LeaderboardSpec } from "./leaderboard";
import type { MarkdownSpec } from "./markdown";
import type { PivotSpec, TableSpec } from "./pivot";

export type ComponentWithMetricsView =
  | CartesianCanvasChartSpec
  | CircularCanvasChartSpec
  | PivotSpec
  | TableSpec
  | KPISpec
  | KPIGridSpec
  | LeaderboardSpec;

export type ComponentSpec = ComponentWithMetricsView | ImageSpec | MarkdownSpec;

export interface ComponentCommonProperties {
  title?: string;
  description?: string;
  show_description_as_tooltip?: boolean;
}

export type VeriticalAlignment = "top" | "middle" | "bottom";
export type HoritzontalAlignment = "left" | "center" | "right";
export interface ComponentAlignment {
  vertical: VeriticalAlignment;
  horizontal: HoritzontalAlignment;
}

export type ComponentComparisonOptions =
  | "previous"
  | "delta"
  | "percent_change";

export interface ComponentFilterProperties {
  time_filters?: string;
  dimension_filters?: string;
}

export interface ComponentSize {
  width: number;
  height: number;
}

export type CanvasComponentType =
  | ChartType
  | "markdown"
  | "kpi_grid"
  | "image"
  | "pivot"
  | "table"
  | "leaderboard";

export type ComponentWithTypeSpec =
  | { line_chart: CartesianCanvasChartSpec }
  | { bar_chart: CartesianCanvasChartSpec }
  | { area_chart: CartesianCanvasChartSpec }
  | { stacked_bar: CartesianCanvasChartSpec }
  | { stacked_bar_normalized: CartesianCanvasChartSpec }
  | { donut_chart: CircularCanvasChartSpec }
  | { funnel_chart: FunnelChartSpec }
  | { heatmap: HeatmapChartSpec }
  | { combo_chart: ComboChartSpec }
  | { pivot: PivotSpec }
  | { table: TableSpec }
  | { kpi_grid: KPIGridSpec }
  | { markdown: MarkdownSpec }
  | { image: ImageSpec };
