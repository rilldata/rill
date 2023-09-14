// Mock apis

import { getBlock } from "../time-dimension-details/util";
import type { PivotColumnSet, PivotConfig, PivotDimension } from "./types";
import { range } from "./util";

function getDimensionCardinality(dim: string) {
  return parseInt(dim.split("Dim").at(1));
}

function getColumnSetSize(set: PivotColumnSet) {
  let colCardinalities = [];
  set.dims.forEach((col) => {
    colCardinalities.push(getDimensionCardinality(col.def));
  });
  colCardinalities.push(set.measures.length);
  return colCardinalities.reduce((acc, curr) => acc * curr, 1);
}

/**
 * TODO
 * - support arbitrary nesting (recursive)
 * - support nest joinType
 * - support expand/collapse states
 *
 */
export function getMetadata(config: PivotConfig) {
  // Row ct
  let rowCardinalities = [];
  config.rowDims.forEach((r) => {
    rowCardinalities.push(getDimensionCardinality(r.def));
  });

  const rowCt = rowCardinalities.reduce((acc, curr) => acc * curr, 1);

  // Col ct
  const colCt = config.colSets.reduce(
    (acc, curr) => acc + getColumnSetSize(curr),
    0
  );

  return {
    rowCt,
    colCt,
  };
}

// Generate combinations for a single dimension
function singleDimCombinations(dimLabel: string, size: number) {
  const combinations = [];
  for (let i = 0; i < size; i++) {
    combinations.push(`${dimLabel}${i}`);
  }
  return combinations;
}

// Recursive function to generate combinations for all dimensions
function allDimCombinations(dims: PivotDimension[], currentIndex: number) {
  if (currentIndex === dims.length) return [[]];

  const currentDim = dims[currentIndex];
  const currentCombinations = singleDimCombinations(
    `${"h".repeat(currentIndex + 1)}`,
    getDimensionCardinality(currentDim.def)
  );
  const nextCombinations = allDimCombinations(dims, currentIndex + 1);

  const result = [];
  for (let combo of currentCombinations) {
    for (let nextCombo of nextCombinations) {
      result.push([combo, ...nextCombo]);
    }
  }
  return result;
}

function generateCombinations(data: PivotColumnSet, headerDepth: number) {
  const dimCombinations = allDimCombinations(data.dims, 0);

  // Now pair every dimension combination with each measure
  const finalCombinations = [];
  for (let dimCombo of dimCombinations) {
    for (let measure of data.measures) {
      const nextColumn = [...dimCombo, measure.def];
      // if(nextColumn.length < headerDepth)
      while (nextColumn.length < headerDepth) {
        nextColumn.unshift("");
      }
      finalCombinations.push(nextColumn);
    }
  }

  return finalCombinations;
}

function getColumnSetHeaders(set: PivotColumnSet, headerDepth: number) {
  return generateCombinations(set, headerDepth);
}

// function getColumnHeaderDepth(set: PivotColumnSet) {

// }

export function getColumnHeaders(config: PivotConfig, x0: number, x1: number) {
  const headerDepth = config.colSets.reduce(
    (acc, curr) => Math.max(curr.dims.length + 1, acc),
    0
  );
  return config.colSets
    .flatMap((set) => getColumnSetHeaders(set, headerDepth))
    .slice(x0, x1);
}

export function getRowHeaders(config: PivotConfig, y0: number, y1: number) {
  const headers = allDimCombinations(config.rowDims, 0);
  // Hack: Add zero width space character to every other line to prevent row merging (for flat tables)
  // see https://github.com/finos/regular-table/issues/193
  for (let i = 0; i < headers.length; i += 2) {
    for (let j = 0; j < headers[i].length; j++) {
      headers[i][j] += "\u200B";
    }
  }
  return headers.slice(y0, y1);
}

const dimNames = ["DimA", "DimB", "DimC", "DimD", "DimE"];
const EXPANDED = [0, 1, 4];
const CUMULATIVE_EXPANDED = EXPANDED.reduce((acc, curr) => {
  const prev = acc.at(-1);
  if (prev === undefined) return [curr];
  else return [...acc, curr + 3 * acc.length];
}, []);

export const fetchMockRowData = (block, delay) => async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        block: block,
        // For nested data, one way to do it is use cumulative_expanded
        // with a for loop, and basically add multiple rows in 1 pass for expanded + jump the index ahead
        // as opposed to do thing this Array 50 method.
        // will need a function to do loop. or maybe flatMap would work as well...just need an early exit OR can just slice array at end
        data: Array.from(Array(50).keys()).map((y) => {
          const parentRow = (block[0] + y) % 4 === 0;
          const rowGroupIdx = Math.floor((block[0] + y) / 4);
          const isExpanded = EXPANDED.includes(rowGroupIdx);
          if (isExpanded) {
            if (parentRow) return [`- ${dimNames[0]}_${rowGroupIdx}`];
            else return ["", `${dimNames[1]}_${block[0] + y}`];
          } else return [`+ ${dimNames[0]}_${block[0] + y}`];
        }),
        //   Array.from(Array(2).keys()).map(
        //     // (x) => `${dimNames[x]}_${x},${block[0] + y}`
        //     (x) => {
        //       const parentRow = (block[0] + y) % 4 === 0;
        //       const rowGroupIdx = Math.floor((block[0] + y) / 4);
        //       if(parentRow) {
        //         if(x === 0) return `${dimNames[x]}_${rowGroupIdx}`;
        //         else return
        //       }
        //       if (x === 0) {
        //         return `${dimNames[x]}_${Math.floor((block[0] + y) / 3)}`;
        //       } else return `${dimNames[x]}_${block[0] + y}`;
        //     }
        //   )
        // ),
      });
    }, delay);
  });
};

export const fetchMockColumnData = (block, config, delay) => async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        block: block,
        data: getColumnHeaders(config, block[0], block[1]),
      });
    }, delay);
  });
};

// TODO: transpose here
export const fetchMockBodyData = (block, delay) => async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        block,
        data: range(block.y[0], block.y[1], (y) =>
          range(block.x[0], block.x[1], (x) => `${x},${y}`)
        ),
      });
    }, delay);
  });
};
