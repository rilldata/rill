import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1MetricsViewFilter,
} from "@rilldata/web-common/runtime-client";
import type { ExpandedState } from "@tanstack/svelte-table";

export interface PivotState {
  active: boolean;
  rows: string[];
  columns: string[];
  expanded: ExpandedState;
  rowJoinType: "flat" | "nest";
  sort: any; // TBD
}

export interface PivotDataRow {
  [key: string]: string | number | PivotDataRow[] | undefined;
  subRows?: PivotDataRow[];
}

/**
 * This is the config that is passed to the pivot data store methods
 */
export interface PivotDataStoreConfig {
  measureNames: string[];
  rowDimensionNames: string[];
  colDimensionNames: string[];
  allMeasures: MetricsViewSpecMeasureV2[];
  allDimensions: MetricsViewSpecDimensionV2[];
  filters: V1MetricsViewFilter;
  pivot: PivotState;
}

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
  rowJoinType: "flat" | "nest";
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
