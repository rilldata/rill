import type { PivotConfig } from "./types";

// Example configs

// TDD config: 1 dimension, 1 measure, 1 pivot dimension + measure
export const tdd: PivotConfig = {
  rowDims: [{ def: "Dimension A" }],
  colSets: [
    // 1 set of 3 columns for the overall measure A, sparkline, percent of total data
    {
      dims: [],
      measures: [
        { def: "Measure A" },
        { def: "Measure A", minichart: true, minichartDimension: "Time" },
        { def: "Measure A Percent of Total" },
      ],
    },
    // 1 set for the pivoted time values
    {
      dims: [{ def: "Time" }],
      measures: [{ def: "Measure A" }],
    },
  ],
  rowJoinType: "flat",
  sort: null,
};

export const basicPivot: PivotConfig = {
  rowDims: [
    {
      def: "Dim100",
    },
    {
      def: "Dim3",
    },
  ],
  colSets: [
    {
      dims: [],
      measures: [{ def: "TotalMeasure" }],
    },
    {
      dims: [
        {
          def: "Dim10",
        },
      ],
      measures: [
        {
          def: "MeasureA",
        },
        {
          def: "MeasureB",
        },
      ],
    },
  ],
  rowJoinType: "flat",
  sort: null,
};
