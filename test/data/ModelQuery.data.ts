import type {TestDataColumns} from "./DataLoader.data";
import type {DataProviderData} from "@adityahegde/typescript-test-utils";

type Args = [string, TestDataColumns];

export const SingleTableQueryColumnsTestData: TestDataColumns = [{
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
export const SingleTableQuery = "select count(*) as impressions, publisher, domain from AdBids group by publisher, domain";
const SingleTableQueryTestData: Args = [SingleTableQuery, SingleTableQueryColumnsTestData];

export const TwoTableJoinQueryColumnsTestData: TestDataColumns = [{
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
export const TwoTableJoinQuery = `
select count(*) as impressions, avg(bid.bid_price) as bid_price, bid.publisher, bid.domain, imp.city, imp.country
from AdBids bid join AdImpressions imp on bid.id = imp.id
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

export const NestedQuery = `
select
    count(*), avg(bid.bid_price) as bid_price,
    bid.publisher, bid.domain, imp.city, imp.country,
    CASE WHEN imp.country = 'India' THEN 'TRUE' ELSE 'FALSE' END as indian
from
    AdBids bid join
    (select imp.id, imp.city, imp.country, u.name from AdImpressions imp join Users u on imp.user_id=u.id where u.city like 'B%') as imp
    on bid.id = imp.id
group by bid.publisher, bid.domain, imp.city, imp.country
`;

export const CTE = `
with
    UserImpression as (
        select
            imp.id, imp.city, imp.country, u.name
        from AdImpressions imp join Users u on imp.user_id=u.id
    )
    select
        count(*) as impressions,
        avg(bid.bid_price) as bid_price,
        bid.publisher, bid.domain, imp.city, imp.country,
        (select uimp.name from UserImpression uimp where uimp.city=imp.city) as users
    from AdBids bid join AdImpressions imp on bid.id = imp.id
    group by bid.publisher, bid.domain, imp.city, imp.country
`
