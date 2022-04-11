import { expectedQueryTree, expectedSelect, expectedColumn, expectedNestedSelect, expectedCTE, expectedSourceTable } from "../utils/expectedQueryTreeFactory";

const AdBidsAliased = expectedSourceTable("adbids", "bid");
const AdImpressionsAliased = expectedSourceTable("adimpressions", "imp");
const UsersAliased = expectedSourceTable("users", "u");

const AdBidsJoinImpressionsColumns = [
    expectedColumn(["bid.bid_price"], "bid_price"),
    expectedColumn(["bid.publisher"]),
    expectedColumn(["bid.domain"]),
    expectedColumn(["imp.city"]),
    expectedColumn(["imp.country"]),
];

export const SingleTableQueryTree = expectedQueryTree(
    expectedSelect(
        [expectedSourceTable("adbids")],
        [
            expectedColumn(["*"], "impressions"),
            expectedColumn(["publisher"]),
            expectedColumn(["domain"]),
        ],
    ),
    [expectedSourceTable("adbids")],
);

export const TwoTableJoinQueryTree = expectedQueryTree(
    expectedSelect(
        [AdBidsAliased, AdImpressionsAliased],
        [
            expectedColumn(["*"], "impressions"),
            ...AdBidsJoinImpressionsColumns,
        ],
    ),
    [AdBidsAliased, AdImpressionsAliased],
);

const NestedImpressionJoinUsersSelect = expectedSelect(
    [AdImpressionsAliased, UsersAliased],
    [
        expectedColumn(["imp.id"]),
        expectedColumn(["imp.city"]),
        expectedColumn(["imp.country"]),
        expectedColumn(["u.name"]),
    ],
);
const NestedImpressionJoinUsers = expectedNestedSelect(
    NestedImpressionJoinUsersSelect,
    "imp",
);
export const NestedQueryTree = expectedQueryTree(
    expectedSelect(
        [AdBidsAliased, NestedImpressionJoinUsers],
        [
            expectedColumn(["*"]),
            ...AdBidsJoinImpressionsColumns,
            expectedColumn(["imp.country"], "indian"),
        ],
    ),
    [AdBidsAliased, AdImpressionsAliased, UsersAliased],
);

const CTEUserImpressions = expectedNestedSelect(
    NestedImpressionJoinUsersSelect,
    "userimpression",
);
export const CTEQueryTree = expectedQueryTree(
    expectedCTE([CTEUserImpressions], expectedSelect(
        [AdBidsAliased, AdImpressionsAliased],
        [
            expectedColumn(["*"], "impressions"),
            ...AdBidsJoinImpressionsColumns,
            // TODO: if there is a requirement parse select as column
            expectedColumn([], "users"),
        ],
    )),
    [
        AdImpressionsAliased, UsersAliased,
        AdBidsAliased, AdImpressionsAliased,
    ],
);
