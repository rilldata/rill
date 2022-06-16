import { expectedQueryTree, expectedSelect, expectedColumn, expectedNestedSelect, expectedCTE, expectedSource } from "../utils/expectedQueryTreeFactory";

const AdBidsAliased = expectedSource("adbids", "bid");
const AdImpressionsAliased = expectedSource("adimpressions", "imp");
const UsersAliased = expectedSource("users", "u");

const AdBidsJoinImpressionsColumns = [
    expectedColumn(["bid.bid_price"], "bid_price"),
    expectedColumn(["bid.publisher"]),
    expectedColumn(["bid.domain"]),
    expectedColumn(["imp.city"]),
    expectedColumn(["imp.country"]),
];

export const SingleSourceQueryTree = expectedQueryTree(
    expectedSelect(
        [expectedSource("adbids")],
        [
            expectedColumn(["*"], "impressions"),
            expectedColumn(["publisher"]),
            expectedColumn(["domain"]),
        ],
    ),
    [expectedSource("adbids")],
);

export const TwoSourceJoinQueryTree = expectedQueryTree(
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
