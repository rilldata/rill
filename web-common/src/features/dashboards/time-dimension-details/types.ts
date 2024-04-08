export enum TDDChart {
  DEFAULT = "default",
  STACKED_BAR = "stacked_bar",
  GROUPED_BAR = "grouped_bar",
  STACKED_AREA = "stacked_area",
}

export type TDDCustomCharts = Exclude<TDDChart, TDDChart.DEFAULT>;
export type TDDBarCharts = Exclude<TDDCustomCharts, TDDChart.STACKED_AREA>;

export interface TDDState {
  /***
   * The name of the measure that is currently being expanded
   * in the Time Detailed Dimension view
   */
  expandedMeasureName?: string;
  /**
   * The index at which selected dimension values are pinned in the
   * time detailed dimension view. Values above this index preserve
   * their original order
   */
  pinIndex: number;
  chartType: TDDChart;
}

export interface HeaderData<T> {
  value: T | null | undefined;
  spark?: string;
}

export type TDDCellData = string | number | null | undefined;

export interface TableData {
  rowCount: number;
  fixedColCount: number;
  rowHeaderData: HeaderData<string>[][];
  columnCount: number;
  columnHeaderData: HeaderData<Date>[][];
  body: TDDCellData[][];
  selectedValues: (string | null)[];
}

export interface HighlightedCell {
  dimensionValue: string | undefined | null;
  time: Date | undefined;
}

export interface ChartInteractionColumns {
  hover: number | undefined;
  scrubStart: number | undefined;
  scrubEnd: number | undefined;
}

export type TDDComparison = "time" | "none" | "dimension";

export interface TablePosition {
  x0?: number;
  x1?: number;
  y0?: number;
  y1?: number;
}
