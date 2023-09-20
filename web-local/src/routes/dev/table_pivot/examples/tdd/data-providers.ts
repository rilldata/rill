import type {
  PivotConfig,
  PivotPos,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import { getColumnHeaders } from "../mock-api";
import { range } from "../util";
import { faker } from "@faker-js/faker";

export const MOCK_ROW_CT = 1000;
export const MOCK_COL_CT = 100;

const sparkTemplate = `<svg width="34" height="13" viewBox="0 0 34 13" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M1 11.75L5.5 5.75L11.5 8.75L17 5.75L21 11.75L28 1.25L33 11.75" stroke="#9CA3AF"/>
</svg>`;

const scale = (n: number) => (n * 13).toFixed(2);
const createSpark = (
  nums: number[]
) => `<svg width="34" height="13" viewBox="0 0 34 13" fill="none" xmlns="http://www.w3.org/2000/svg">
<path d="M1 ${scale(nums[0])}L5.5 ${scale(nums[1])}L11.5 ${scale(
  nums[2]
)}L17 ${scale(nums[3])}L21 ${scale(nums[4])}L28 ${scale(nums[5])}L33 ${scale(
  nums[6]
)}" stroke="#9CA3AF"/>
</svg>`;

// When using row headers, be careful not to accidentally merge cells
const rowHeaderData = range(0, MOCK_ROW_CT, (i) => [
  // Dim
  {
    value: faker.commerce.productName(),
  },
  // Measure total
  {
    value:
      faker.commerce.price({ min: 10, max: 100, dec: 0, symbol: "$" }) +
      "." +
      faker.commerce.price({ min: 10, max: 99, dec: 0 }),
    spark: createSpark(range(0, 7, (i) => Math.random())),
  },
  // Measure percent of total
  {
    value: faker.number.int({ min: 10, max: 99 }) / 10 + "%",
  },
]);

const MOCK_START_DATE = new Date("2023-03-29");
const columnHeaderData = range(0, MOCK_COL_CT, (i) => {
  const d = new Date(MOCK_START_DATE);
  d.setDate(d.getDate() + i);
  return [
    {
      value: d,
    },
    // `Col_${i}`,
  ];
});

export function getRowHeaderData(pos: PivotPos) {
  return rowHeaderData.slice(pos.y0, pos.y1);
}

export function getColumnHeaderData(pos: PivotPos) {
  return columnHeaderData.slice(pos.x0, pos.x1);
}

export function createColumnHeaderDataGetter(config: PivotConfig) {
  return function getColumnHeaderData(pos: PivotPos) {
    return getColumnHeaders(config, pos.x0, pos.x1);
  };
}

const MOCK_BODY = range(0, MOCK_COL_CT, (y) =>
  range(
    0,
    MOCK_ROW_CT,
    (x) =>
      faker.commerce.price({ min: 1, max: 4, dec: 0, symbol: "" }) +
      "." +
      faker.commerce.price({ min: 10, max: 99, dec: 0 })
  )
);

export function getBodyData(pos: PivotPos) {
  /* 
    Important: regular-table expects body data in columnar format,
    aka an array of arrays where outer array is the columns,
    inner array is the row values for a specific column
  */
  return MOCK_BODY.slice(pos.x0, pos.x1).map((row) =>
    row.slice(pos.y0, pos.y1)
  );
}
