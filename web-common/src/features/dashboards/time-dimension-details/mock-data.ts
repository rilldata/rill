const FIXED_COL_CT = 5;

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
    },
    {
      title: "Δ",
    },
    {
      title: "Δ%",
    },
  ],
  metadata: {
    rowCt: 1000,
    fixedColumnCt: FIXED_COL_CT,
    pivotColumnCt: 100,
  },
};

// Populate mock data
for (let r = 0; r < data.metadata.rowCt; r++) {
  const row = new Array(
    data.metadata.fixedColumnCt + data.metadata.pivotColumnCt
  ).fill(0);

  // Fill in fixed columns
  for (let i = 0; i < data.metadata.fixedColumnCt; i++) {
    const cell: { d: any } = { d: "" };
    if (i === 0) {
      cell.d = `Value A${r}`;
    }
    if (i === 1) {
      cell.d = (Math.random() * 1000).toFixed(2);
      cell.spark = [10, 30, 20, 50, 30, 60, 80, 100, 70];
    }
    if (i === 2) {
      cell.d = (Math.random() * 100).toFixed(2) + "%";
    }
    if (i === 3) {
      cell.d = "$" + (Math.random() * 10).toFixed(2);
    }
    if (i === 4) {
      cell.d = (Math.random() * 100 - 50).toFixed(2) + "%";
    }

    row[i] = cell;
  }

  for (let i = data.metadata.fixedColumnCt; i < row.length; i++) {
    row[i] = {
      d: (Math.random() * 10).toFixed(2),
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
