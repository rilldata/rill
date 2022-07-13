import type { DataProviderData } from "@adityahegde/typescript-test-utils";
import {
  AD_BID_COUNT,
  AD_IMPRESSION_COUNT,
  MAX_USERS,
} from "./generator/data-constants";

export type TestDataColumn = {
  name: string;
  type: string;
  isNull: boolean;
};
export type TestDataColumns = Array<TestDataColumn>;
type Args = [string, number, TestDataColumns];

export const AdBidsColumnsTestData: TestDataColumns = [
  {
    name: "id",
    type: "INTEGER",
    isNull: false,
  },
  {
    name: "timestamp",
    type: "TIMESTAMP",
    isNull: false,
  },
  {
    name: "publisher",
    type: "VARCHAR",
    isNull: true,
  },
  {
    name: "domain",
    type: "VARCHAR",
    isNull: false,
  },
  {
    name: "bid_price",
    type: "DOUBLE",
    isNull: false,
  },
];

export const AdImpressionColumnsTestData: TestDataColumns = [
  {
    name: "id",
    type: "INTEGER",
    isNull: false,
  },
  {
    name: "city",
    type: "VARCHAR",
    isNull: true,
  },
  {
    name: "country",
    type: "VARCHAR",
    isNull: false,
  },
  {
    name: "user_id",
    type: "INTEGER",
    isNull: true,
  },
];

export const UserColumnsTestData: TestDataColumns = [
  {
    name: "id",
    type: "INTEGER",
    isNull: false,
  },
  {
    name: "name",
    type: "VARCHAR",
    isNull: false,
  },
  {
    name: "city",
    type: "VARCHAR",
    isNull: true,
  },
  {
    name: "country",
    type: "VARCHAR",
    isNull: false,
  },
];

export const normaliseParquetColumn = (testDataColumns: TestDataColumns) => {
  // Parquet generator is not generating timestamp. it is instead loaded as varchar
  return testDataColumns.map((testDataColumn) => {
    return {
      ...testDataColumn,
      type:
        testDataColumn.type === "TIMESTAMP" ? "VARCHAR" : testDataColumn.type,
    };
  });
};

export type FileImportTestDataProvider = DataProviderData<Args>;
export const ParquetFileTestData: FileImportTestDataProvider = {
  title: "ParquetFiles",
  subData: [
    ["AdBids.parquet", AD_BID_COUNT, AdBidsColumnsTestData],
    ["AdImpressions.parquet", AD_IMPRESSION_COUNT, AdImpressionColumnsTestData],
    ["Users.parquet", MAX_USERS, UserColumnsTestData],
  ].map((data: Args) => {
    return {
      title: data[0],
      args: [data[0], data[1], normaliseParquetColumn(data[2])],
    };
  }),
};
export const CSVFileTestData: FileImportTestDataProvider = {
  title: "CSVFiles",
  subData: [
    ["AdBids.csv", AD_BID_COUNT, AdBidsColumnsTestData],
    ["AdImpressions.tsv", AD_IMPRESSION_COUNT, AdImpressionColumnsTestData],
    ["Users.csv", MAX_USERS, UserColumnsTestData],
  ].map((data: Args) => {
    return {
      title: data[0],
      args: data,
    };
  }),
};
