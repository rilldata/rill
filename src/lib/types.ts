/**
 * The type definition for a "profile column"
 */
export interface ProfileColumn {
  name: string;
  type: string;
  conceptualType: string;
  summary?: ProfileColumnSummary | any;
  nullCount?: number;
}

export interface Item {
  id: string;
  status?: string;
}

export interface ColumnarItem extends Item {
  profile?: ProfileColumn[]; // TODO: create Profile interface
}

export interface QueryElementReference {
  name: string;
  start: number;
  end: number;
}

export interface Source {
  name: string;
  sources: QueryElementReference[];
}

export interface Model extends ColumnarItem {
  /**  */
  query: string;
  /** sanitizedQuery is always a 1:1 function of the query itself */
  sanitizedQuery: string;
  /** name is used for the filename and exported file */
  name: string;
  sourceName?: string;
  /** cardinality is the total number of rows of the previewed table */
  cardinality?: number;
  /** sizeInBytes is the total size of the previewed table.
   * It is not generated until the user exports the query.
   * */
  sizeInBytes?: number; // TODO: make sure this is just size
  error?: string;
  sources?: Source[];
  preview?: any;
  destinationProfile?: any;
}

export enum SourceType {
  ParquetFile,
  CSVFile,
  // source is loaded from an existing duckdb table
  DuckDB,
}
export const FILE_EXTENSION_TO_SOURCE_TYPE = {
  parquet: SourceType.ParquetFile,
  csv: SourceType.CSVFile,
  tsv: SourceType.CSVFile,
};

export type ProfileColumnSummary =
  | CategoricalSummary
  | NumericSummary
  | TimeRangeSummary;

export interface CategoricalSummary {
  topK: TopKEntry[];
  cardinality: number;
}

export interface NumericSummary {
  histogram?: NumericHistogramBin[];
  statistics?: NumericStatistics;
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

export interface TimeRangeSummary {
  min: string;
  max: string;
  interval: {
    months: number;
    days: number;
    micros: number;
  };
}

/** Metrics Models
 * A metrics model
 */

export interface MetricsModel {
  /** the current materialized table */
  id: string;
  name: string;
  spec: string;
  parsedSpec?: any;
  error?: string;
  // table?: string;
  // timeField?: string;
  // timeGrain?: 'day' | 'hour';
  // metrics: MetricConfiguration[];
  // dimensions: DimensionConfiguration[];
  // activeMetrics:string[];
  // activeDimensions:string[];
  // view: MetricsModelView;
}

export interface MetricsModelView {
  metrics: TimeSeries[];
  dimensions: Leaderboard[];
}

export interface TimeSeries {
  name: string;
  data: any[];
}

export interface Leaderboard {
  name: string;
  data: any[];
}

export interface MetricConfiguration {
  name: string;
  transform: string;
  field: string;
  description?: string;
  id: string;
}

export interface DimensionConfiguration {
  name: string;
  field: string;
  id: string;
}

export interface Asset {
  id: string;
  assetType: string;
}

export interface ExploreConfiguration {
  id: string;
  modelID: string;
  name: string;
  activeMetrics: string[];
  activeDimensions: string[];
  selectedMetrics: string[];
  selectedDimensions: string[];
  currentMetricLeaderboard?: string;
  preview: any;
}

/**
 * The entire state object for the Rill Developer.
 */
export interface DataModelerState {
  activeAsset?: Asset;
  models: Model[];
  sources: Source[];
  metricsModels: MetricsModel[];
  exploreConfigurations: ExploreConfiguration[];
  status: string;
}

export type ColumnarTypeKeys = {
  [K in keyof DataModelerState]: DataModelerState[K] extends Model[] | Source[]
    ? K
    : never;
}[keyof DataModelerState];
