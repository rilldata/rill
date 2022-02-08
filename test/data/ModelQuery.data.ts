import type {TestDataColumns} from "./DataLoader.data";
import type {DataProviderData} from "@adityahegde/typescript-test-utils";

type Args = [string, TestDataColumns];

const SingleTableQueryColumnsTestData: TestDataColumns = [{
    name: "impressions",
    type: "BIGINT",
    isNull: false,
}, {
    name: "publisher",
    type: "VARCHAR",
    isNull: true,
}, {
    name: "domain",
    type: "VARCHAR",
    isNull: false,
}];
const SingleTableQueryTestData: Args = [
    "select count(*) as impressions, publisher, domain from 'AdBids.parquet' group by publisher, domain",
    SingleTableQueryColumnsTestData,
];

const TwoTableJoinQueryColumnsTestData: TestDataColumns = [{
    name: "impressions",
    type: "BIGINT",
    isNull: false,
}, {
    name: "bid_price",
    type: "DOUBLE",
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
    name: "city",
    type: "VARCHAR",
    isNull: true,
}, {
    name: "country",
    type: "VARCHAR",
    isNull: false,
}];
const TwoTableJoinQueryTestData: Args = [
    `
        select count(*) as impressions, avg(bid.bid_price) as bid_price, bid.publisher, bid.domain, imp.city, imp.country
        from 'AdBids.parquet' bid join 'AdImpressions.parquet' imp on bid.id = imp.id
        group by bid.publisher, bid.domain, imp.city, imp.country
    `,
    TwoTableJoinQueryColumnsTestData,
];

export type QueryInfoTestDataProvider = DataProviderData<Args>;
export const QueryInfoTestData: QueryInfoTestDataProvider = {
    subData: [{
        title: "Single table group",
        args: SingleTableQueryTestData,
    }, {
        title: "Two table join",
        args: TwoTableJoinQueryTestData,
    }],
};
