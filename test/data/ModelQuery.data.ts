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
export const SingleTableQuery = "select count(*) as impressions, publisher, domain from 'AdBids_parquet' group by publisher, domain";
const SingleTableQueryTestData: Args = [SingleTableQuery, SingleTableQueryColumnsTestData];

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
const TwoTableJoinQuery = `
select count(*) as impressions, avg(bid.bid_price) as bid_price, bid.publisher, bid.domain, imp.city, imp.country
from 'AdBids_parquet' bid join 'AdImpressions_parquet' imp on bid.id = imp.id
group by bid.publisher, bid.domain, imp.city, imp.country
`;
const TwoTableJoinQueryTestData: Args = [TwoTableJoinQuery, TwoTableJoinQueryColumnsTestData];

export type ModelQueryTestDataProvider = DataProviderData<Args>;
export const ModelQueryTestData: ModelQueryTestDataProvider = {
    subData: [{
        title: "Single table group",
        args: SingleTableQueryTestData,
    }, {
        title: "Two table join",
        args: TwoTableJoinQueryTestData,
    }],
};
