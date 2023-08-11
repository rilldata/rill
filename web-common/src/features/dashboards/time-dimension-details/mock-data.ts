const FIXED_COL_CT = 6;

export const data = {
  data: [],
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
  )
    .fill(0)
    .map((d, i) => ({
      d: `cell ${r},${i}`,
    }));

  data.data.push(row);
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
