import { expectedQueryTree, expectedSelect, expectedTable, expectedColumn, expectedNestedSelect, expectedCTE } from "../utils/expectedQueryTreeFactory";

const AdBidsAliased = expectedTable("adbids", "bid");
const AdImpressionsAliased = expectedTable("adimpressions", "imp");
const UsersAliased = expectedTable("users", "u");

const AdBidsJoinImpressionsColumns = [
    expectedColumn(["bid.bid_price"], "bid_price"),
    expectedColumn(["bid.publisher"]),
    expectedColumn(["bid.domain"]),
    expectedColumn(["imp.city"]),
    expectedColumn(["imp.country"]),
];

export const SingleTableQueryTree = expectedQueryTree(
    expectedSelect(
        [expectedTable("adbids")],
        [
            expectedColumn(["*"], "impressions"),
            expectedColumn(["publisher"]),
            expectedColumn(["domain"]),
        ],
    ),
    [expectedTable("adbids")],
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
    [AdBidsAliased, NestedImpressionJoinUsers, AdImpressionsAliased, UsersAliased],
);

const CTEUserImpressions = expectedNestedSelect(
    NestedImpressionJoinUsersSelect,
    "userimpression",
);
const UserImpressionAliased = expectedTable("userimpression", "uimp");
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
        CTEUserImpressions, AdImpressionsAliased, UsersAliased,
        AdBidsAliased, AdImpressionsAliased, UserImpressionAliased,
    ],
);
