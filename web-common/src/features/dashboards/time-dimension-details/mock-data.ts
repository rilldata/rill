import Delta from "@rilldata/web-common/components/icons/Delta.svelte";
import DeltaChange from "../dimension-table/DeltaChange.svelte";
import PercentOfTotal from "../dimension-table/PercentOfTotal.svelte";
import DeltaChangePercentage from "../dimension-table/DeltaChangePercentage.svelte";
import type { SvelteComponent } from "svelte";

const FIXED_COL_CT = 5;

export type THeaderData = {
  title?: string;
  component?: typeof SvelteComponent;
};

export const data = {
  data: [],
  headers: [
    {
      title: "Dimension A",
    },
    {
      title: "Measure A",
    },
    {
      title: "% total",
      component: PercentOfTotal,
    },
    {
      title: "Δ",
      component: Delta,
    },
    {
      title: "Δ%",
      component: DeltaChangePercentage,
    },
  ],
  metadata: {
    rowCt: 1000,
    fixedColumnCt: FIXED_COL_CT,
    pivotColumnCt: 100,
  },
};

export type TCellData = {
  text?: string;
  value?: number;
  spark?: number[];
};

// Populate mock data
for (let r = 0; r < data.metadata.rowCt; r++) {
  const row = new Array<TCellData>(
    data.metadata.fixedColumnCt + data.metadata.pivotColumnCt
  ).fill(undefined);

  // Fill in fixed columns
  for (let i = 0; i < data.metadata.fixedColumnCt; i++) {
    const cell: { text?: string; value?: number; spark?: number[] } = {};
    if (i === 0) {
      cell.text = `Value A${r}`;
    }
    // Total and spark line
    if (i === 1) {
      cell.value = Math.random() * 1000;
      cell.spark = [10, 30, 20, 50, 30, 60, 80, 100, 70];
    }
    // % of total
    if (i === 2) {
      cell.value = Math.random();
    }
    // Delta
    if (i === 3) {
      cell.value = Math.random() * 10 - 5;
    }
    // Delta %
    if (i === 4) {
      cell.value = Math.random() - 0.5;
    }

    row[i] = cell;
  }

  for (let i = data.metadata.fixedColumnCt; i < row.length; i++) {
    row[i] = {
      value: Math.random() * 10,
    };
  }

  data.data.push(row);
}

const shortFormatDate = new Intl.DateTimeFormat(undefined, {
  month: "short",
  day: "numeric",
}).format;

// Populate column headers
const startDate = new Date("3/29/2023");
for (let i = 0; i < data.metadata.pivotColumnCt; i++) {
  const date = new Date(startDate);
  date.setDate(date.getDate() + i);
  data.headers.push({
    title: shortFormatDate(date),
  });
}

// Mock data fetch
export const fetchData = (block, delay) => async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        data: data.data.slice(block[0], block[1]),
        block,
      });
    }, delay);
  });
};
