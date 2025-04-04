import { type TimeRangeString } from "@rilldata/web-common/lib/time/types";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1Expression,
  V1MetricsViewAggregationResponseDataItem,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";
import type {
  ColumnDef,
  ExpandedState,
  SortingState,
} from "@tanstack/svelte-table";
import type { Readable } from "svelte/motion";

export const COMPARISON_VALUE = "__comparison_value";
export const COMPARISON_DELTA = "__comparison_delta";
export const COMPARISON_PERCENT = "__comparison_percent";

export interface PivotDataState {
  isFetching: boolean;
  error?: PivotQueryError[];
  data: PivotDataRow[];
  columnDef: ColumnDef<PivotDataRow>[];
  assembled: boolean;
  totalColumns: number; // total columns excluding row and group totals columns
  reachedEndForRowData?: boolean;
  totalsRowData?: PivotDataRow;
  activeCellFilters?: PivotFilter;
}

export type PivotDataStore = Readable<PivotDataState>;

export interface PivotCell {
  rowId: string;
  columnId: string;
}

export interface PivotDashboardContext {
  metricsViewName: Readable<string>;
  queryClient: QueryClient;
  enabled: boolean;
}

export interface PivotState {
  active: boolean;
  columns: PivotChipData[];
  rows: PivotChipData[];
  expanded: ExpandedState;
  sorting: SortingState;
  columnPage: number;
  rowPage: number;
  enableComparison: boolean;
  tableMode: PivotTableMode;
  activeCell: PivotCell | null;
}

export type PivotTableMode = "flat" | "nest";

export interface PivotDataRow {
  subRows?: PivotDataRow[];

  [key: string]: string | number | PivotDataRow[] | undefined;
}

export interface TimeFilters {
  timeStart: string;
  interval: V1TimeGrain;
  // Time end represents the start time of the last interval for a range
  timeEnd?: string;
}

export interface PivotTimeConfig {
  timeStart: string | undefined;
  timeEnd: string | undefined;
  timeZone: string;
  timeDimension: string;
}

export interface PivotQueryError {
  statusCode: number | null;
  message?: string;
}

/**
 * This is the config that is passed to the pivot data store methods
 */
export interface PivotDataStoreConfig {
  measureNames: string[];
  rowDimensionNames: string[];
  colDimensionNames: string[];
  allMeasures: MetricsViewSpecMeasure[];
  allDimensions: MetricsViewSpecDimension[];
  whereFilter: V1Expression;
  pivot: PivotState;
  time: PivotTimeConfig;
  enableComparison: boolean;
  comparisonTime: TimeRangeString | undefined;
  searchText: string | undefined;
  isFlat: boolean;
}

export interface PivotAxesData {
  isFetching: boolean;
  data?: Record<string, string[]> | undefined;
  totals?:
    | Record<string, V1MetricsViewAggregationResponseDataItem[]>
    | undefined;
  error?: PivotQueryError[];
}

export interface PivotFilter {
  filters: V1Expression | undefined;
  timeRange: TimeRangeString;
}

// OLD PIVOT TYPES
export type PivotMeasure = {
  def: string;
  minichart?: boolean;
  minichartDimension?: string;
  /* expand with other props over time as needed */
};

export type PivotDimension = {
  def: string;
  /* other props like sort criteria, limits can go here */
};

export type PivotColumnSet = {
  dims: PivotDimension[];
  measures: PivotMeasure[];
};

export type PivotConfig = {
  rowDims: PivotDimension[];
  colSets: PivotColumnSet[];
  sort: any; // TBD
  expanded: any[];
};

export type PivotPos = {
  x0: number;
  x1: number;
  y0: number;
  y1: number;
};

export type PivotRenderCallback = (data: {
  x: number;
  y: number;
  value: any;
  element: HTMLElement;
}) => string | void;

export type PivotSidebarSection = "Time" | "Measures" | "Dimensions";

export type PivotChipData = {
  id: string;
  title: string;
  type: PivotChipType;
  description?: string;
};

export enum PivotChipType {
  Time = "time",
  Measure = "measure",
  Dimension = "dimension",
}

export type MeasureType = "measure" | "comparison_delta" | "comparison_percent";
