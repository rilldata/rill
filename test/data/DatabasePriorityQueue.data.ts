export const AdBidsImportActions = [
    [ "importParquetFile", [ "data/AdBids.parquet", "AdBids" ] ],
    [
        "validateQuery",
        [
            "select count(*) as impressions, publisher, domain from 'AdBids' group by publisher, domain"
        ]
    ],
    [ "getProfileColumns", [ "AdBids" ] ],
    [ "getDestinationSize", [ "data/AdBids.parquet" ] ],
    [ "getCardinalityOfTable", [ "AdBids" ] ],
    [ "getFirstNOfTable", [ "AdBids" ] ],
];
export const AdBidsProfilingActions = [
    [ "getNullCount", [ "AdBids", "id" ] ],
    [ "getNumericHistogram", [ "AdBids", "timestamp", "TIMESTAMP" ] ],
    [ "getTimeRange", [ "AdBids", "timestamp" ] ],
    [ "getNullCount", [ "AdBids", "timestamp" ] ],
    [ "getTopKAndCardinality", [ "AdBids", "publisher" ] ],
    [ "getNullCount", [ "AdBids", "publisher" ] ],
    [ "getTopKAndCardinality", [ "AdBids", "domain" ] ],
    [ "getNullCount", [ "AdBids", "domain" ] ],
    [ "getNumericHistogram", [ "AdBids", "bid_price", "DOUBLE" ] ],
    [ "getDescriptiveStatistics", [ "AdBids", "bid_price" ] ],
    [ "getNullCount", [ "AdBids", "bid_price" ] ],
];

export const SingleQueryProfilingActions = [
    [ "getNumericHistogram", [ "query_0", "impressions", "BIGINT" ] ],
    [ "getDescriptiveStatistics", [ "query_0", "impressions" ] ],
    [ "getNullCount", [ "query_0", "impressions" ] ],
    [ "getTopKAndCardinality", [ "query_0", "publisher" ] ],
    [ "getNullCount", [ "query_0", "publisher" ] ],
    [ "getTopKAndCardinality", [ "query_0", "domain" ] ],
    [ "getNullCount", [ "query_0", "domain" ] ],
    [ "getFirstNOfTable", [ "query_0", 25 ] ],
    [ "getCardinalityOfTable", [ "query_0" ] ],
    [ "getDestinationSize", [ "query_0" ] ],
];
export const TwoTableJoinQueryProfilingActions = [
    [ "getDescriptiveStatistics", [ "query_1", "bid_price" ] ],
    [ "getNullCount", [ "query_1", "bid_price" ] ],
    [ "getTopKAndCardinality", [ "query_1", "publisher" ] ],
    [ "getNullCount", [ "query_1", "publisher" ] ],
    [ "getTopKAndCardinality", [ "query_1", "domain" ] ],
    [ "getNullCount", [ "query_1", "domain" ] ],
    [ "getTopKAndCardinality", [ "query_1", "city" ] ],
    [ "getNullCount", [ "query_1", "city" ] ],
    [ "getTopKAndCardinality", [ "query_1", "country" ] ],
    [ "getNullCount", [ "query_1", "country" ] ],
    [ "getFirstNOfTable", [ "query_1", 25 ] ],
    [ "getCardinalityOfTable", [ "query_1" ] ],
    [ "getDestinationSize", [ "query_1" ] ]
]
