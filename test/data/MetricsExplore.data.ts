import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { PreviewRollupInterval } from "$lib/duckdb-data-types";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import {
  getTimeRange,
  TimeSeriesMeasureRange,
} from "../utils/time-series-helpers";

export interface MetricsExploreTestDataType {
  title: string;

  // request arguments
  measures?: Array<number>;
  filters?: ActiveValues;
  timeRange?: TimeSeriesTimeRange;

  // assert arguments
  previewRollupInterval: PreviewRollupInterval;
  measureRanges?: Array<TimeSeriesMeasureRange>;
  leaderboards: Array<[string, Array<string>]>;
  bigNumber: TimeSeriesMeasureRange;
}

const DefaultDomainLeaderboard: [string, Array<string>] = [
  "domain",
  [
    "facebook.com",
    "google.com",
    "msn.com",
    "news.yahoo.com",
    "instagram.com",
    "news.google.com",
    "sports.yahoo.com",
  ],
];
const DefaultCityLeaderboard: [string, Array<string>] = [
  "city",
  [
    "Bengaluru",
    "Boston",
    "Delhi",
    "Dublin",
    "Kolkata",
    "London",
    "Los Angeles",
    "Mumbai",
    "New York City",
    "San Francisco",
  ],
];

export const MetricsExploreTestData: Array<MetricsExploreTestDataType> = [
  {
    title: "basic request",
    previewRollupInterval: PreviewRollupInterval.day,
    leaderboards: [
      ["publisher", ["Facebook", "Google", "Yahoo", "Microsoft"]],
      DefaultDomainLeaderboard,
      DefaultCityLeaderboard,
      ["country", ["India", "USA", "Ireland", "UK"]],
    ],
    bigNumber: { impressions: [50000, 50000], bid_price: [2.75, 3.25] },
  },
  {
    title: "request by month",
    timeRange: getTimeRange(PreviewRollupInterval.month),
    previewRollupInterval: PreviewRollupInterval.month,
    leaderboards: [
      ["publisher", ["Facebook", "Google", "Yahoo", "Microsoft"]],
      DefaultDomainLeaderboard,
      DefaultCityLeaderboard,
      ["country", ["India", "USA", "Ireland", "UK"]],
    ],
    bigNumber: { impressions: [50000, 50000], bid_price: [2.75, 3.25] },
  },
  {
    title: "request with filters",
    timeRange: getTimeRange(PreviewRollupInterval.month),
    previewRollupInterval: PreviewRollupInterval.month,
    filters: {
      domain: [["sports.yahoo.com", true]],
    },
    measureRanges: [
      { impressions: [3500, 4500], bid_price: [3, 4] },
      { impressions: [750, 1250], bid_price: [1, 2] },
      { impressions: [750, 1250], bid_price: [1, 2] },
    ],
    leaderboards: [
      ["publisher", ["Yahoo"]],
      ["domain", ["sports.yahoo.com"]],
      DefaultCityLeaderboard,
      ["country", ["India", "USA", "Ireland", "UK"]],
    ],
    bigNumber: { impressions: [6000, 7000], bid_price: [2.5, 3] },
  },
  {
    title: "request with filters and time range and single measure",
    timeRange: getTimeRange(PreviewRollupInterval.month, "2022-02-01"),
    previewRollupInterval: PreviewRollupInterval.month,
    filters: {
      publisher: [["Yahoo", false]],
    },
    measures: [1],
    measureRanges: [{ bid_price: [2.5, 3] }, { bid_price: [3, 3.5] }],
    leaderboards: [
      ["publisher", ["Microsoft", "Facebook", "Google"]],
      [
        "domain",
        [
          "facebook.com",
          "google.com",
          "msn.com",
          "instagram.com",
          "news.google.com",
        ],
      ],
      DefaultCityLeaderboard,
      ["country", ["India", "Ireland", "UK", "USA"]],
    ],
    bigNumber: { bid_price: [2.75, 3.25] },
  },
  {
    title: "request with missing start time",
    timeRange: getTimeRange(
      PreviewRollupInterval.month,
      "2021-11-01",
      "2022-02-28"
    ),
    previewRollupInterval: PreviewRollupInterval.month,
    measures: [1],
    measureRanges: [
      { bid_price: [0, 0] },
      { bid_price: [0, 0] },
      { bid_price: [2.75, 3.25] },
      { bid_price: [2.75, 3.25] },
    ],
    leaderboards: [
      ["publisher", ["Google", "Yahoo", "Facebook", "Microsoft"]],
      [
        "domain",
        [
          "google.com",
          "instagram.com",
          "news.google.com",
          "news.yahoo.com",
          "sports.yahoo.com",
          "facebook.com",
          "msn.com",
        ],
      ],
      DefaultCityLeaderboard,
      ["country", ["India", "Ireland", "UK", "USA"]],
    ],
    bigNumber: { bid_price: [2.75, 3.25] },
  },
];
