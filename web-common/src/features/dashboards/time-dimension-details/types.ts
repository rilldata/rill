export interface HeaderData {
  value: string | null | undefined;
  spark?: string;
}

export type TDDCellData = string | number | null | undefined;

export interface TableData {
  rowCount: number;
  fixedColCount: number;
  rowHeaderData: HeaderData[][];
  columnCount: number;
  columnHeaderData: HeaderData[][];
  body: TDDCellData[][];
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
