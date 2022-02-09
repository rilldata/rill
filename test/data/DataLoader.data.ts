import type {DataProviderData} from "@adityahegde/typescript-test-utils";
import {AD_BID_COUNT, AD_IMPRESSION_COUNT, MAX_USERS} from "./generator/data-constants";

export type TestDataColumn = {
    name: string;
    type: string;
    isNull: boolean;
};
export type TestDataColumns = Array<TestDataColumn>;
type Args = [string, number, TestDataColumns];

const AdBidsColumnsTestData: TestDataColumns = [{
    name: "id",
    type: "BIGINT",
    isNull: false,
}, {
    name: "timestamp",
    type: "TIMESTAMP",
    isNull: false,
}, {
    name: "publisher",
    type: "VARCHAR",
    isNull: true,
}, {
    name: "domain",
    type: "VARCHAR",
    isNull: false,
}, {
    name: "bid_price",
    type: "DOUBLE",
    isNull: false,
}];
const AdBidsTestData: Args = [
    "AdBids.parquet",
    AD_BID_COUNT,
    AdBidsColumnsTestData,
];

const AdImpressionColumnsTestData: TestDataColumns = [{
    name: "id",
    type: "BIGINT",
    isNull: false,
}, {
    name: "city",
    type: "VARCHAR",
    isNull: true,
}, {
    name: "country",
    type: "VARCHAR",
    isNull: false,
}, {
    name: "user_id",
    type: "BIGINT",
    isNull: true,
}];
const AdImpressionTestData: Args = [
    "AdImpressions.parquet",
    AD_IMPRESSION_COUNT,
    AdImpressionColumnsTestData,
];

const UserColumnsTestData: TestDataColumns = [{
    name: "id",
    type: "BIGINT",
    isNull: false,
}, {
    name: "name",
    type: "VARCHAR",
    isNull: false,
}, {
    name: "city",
    type: "VARCHAR",
    isNull: true,
}, {
    name: "country",
    type: "VARCHAR",
    isNull: false,
}];
const UserTestData: Args = [
    "Users.parquet",
    MAX_USERS,
    UserColumnsTestData,
];

export type ParquetFileTestDataProvider = DataProviderData<Args>;
export const ParquetFileTestData: ParquetFileTestDataProvider = {
    subData: [
        AdBidsTestData, AdImpressionTestData, UserTestData
    ].map(data => {
        return { title: data[0], args: data };
    }),
};
