import type {
  PivotConfig,
  PivotPos,
} from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  getColumnHeaders,
  getFlatRowHeaders,
  getNestedRowHeaders,
} from "../mock-api";
import { range } from "../util";

export function createRowHeaderDataGetter(config: PivotConfig) {
  return function getRowHeaderData(pos: PivotPos) {
    return config.rowJoinType === "flat"
      ? getFlatRowHeaders(config, pos.y0, pos.y1)
      : getNestedRowHeaders(config, pos.y0, pos.y1);
  };
}

export function createColumnHeaderDataGetter(config: PivotConfig) {
  return function getColumnHeaderData(pos: PivotPos) {
    return getColumnHeaders(config, pos.x0, pos.x1);
  };
}

const MOCK_BODY_DATA = range(0, 1000, (x) =>
  range(0, 1000, (y) => "$" + (Math.random() * 10).toFixed(2))
);

export function getBodyData(pos: PivotPos) {
  /* 
    Important: regular-table expects body data in columnar format,
    aka an array of arrays where outer array is the columns,
    inner array is the row values for a specific column
  */
  // return range(pos.x0, pos.x1, (y) =>
  //   range(pos.y0, pos.y1, (x) => `${x},${y}`)
  // );
  return MOCK_BODY_DATA.slice(pos.x0, pos.x1).map((row) =>
    row.slice(pos.y0, pos.y1)
  );
}
