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
  dimensionValue: string;
  time: Date;
}

export interface ChartInteractionColumns {
  hover: number;
  scrubStart: number;
  scrubEnd: number;
}
