export interface ProfileColumn {
  name: string;
  type: string;
  conceptualType: string;
  largestStringLength?: number;
  summary?: ProfileColumnSummary | any;
  nullCount?: number;
}

export interface VirtualizedTableColumns extends ProfileColumn {
  label?: string;
  total?: number;
  description?: string;
  enableResize?: boolean;
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

/** The LayoutElement is the state needed for different layout surfaces,
 * such as the navigation menu, the inspector, and the model output.
 */
export interface LayoutElement {
  value: number;
  visible: boolean;
}
