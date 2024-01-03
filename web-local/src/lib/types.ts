import type { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { SvelteComponent } from "svelte";

export interface VirtualizedTableColumns {
  name: string;
  type: string;
  largestStringLength?: number;
  summary?: ProfileColumnSummary | any;
  label?: string | typeof SvelteComponent<any>;
  total?: number;
  description?: string;
  enableResize?: boolean;
  // is this column highlighted in the table
  highlight?: boolean;
  // Is this the table sorted by this column, and if so, in what direction?
  // Leave undefined if the table is not sorted by this column.
  sorted?: SortDirection;
  format?: string;
}

export type ProfileColumnSummary =
  | CategoricalSummary
  | NumericSummary
  | TimeRangeSummary;

export interface CategoricalSummary {
  topK?: TopKEntry[];
  cardinality?: number;
}

export interface NumericSummary {
  histogram?: NumericHistogramBin[];
  statistics?: NumericStatistics;
  outliers?: NumericOutliers[];
}

export interface TopKEntry {
  value: any;
  count: number;
}

export interface NumericHistogramBin {
  bucket: number;
  low: number;
  high: number;
  count: number;
}

export interface NumericStatistics {
  min: number;
  max: number;
  mean: number;
  q25: number;
  q50: number;
  q75: number;
  sd: number;
}

export interface NumericOutliers {
  bucket: number;
  low: number;
  high: number;
  present: boolean;
}

export interface TimeRangeSummary {
  min: string;
  max: string;
  interval: {
    months: number;
    days: number;
    micros: number;
  };
}
