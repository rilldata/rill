export interface TableData {
  rowCount: number;
  fixedColCount: number;
  rowHeaderData: Array<Array<{ value: string }>>;
  columnCount: number;
  columnHeaderData: Array<Array<{ value: string }>>;
  body: Array<Array<string | number | null>>;
  selectedValues: string[];
}

export interface HighlightedCell {
  dimensionValue: string | undefined;
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
