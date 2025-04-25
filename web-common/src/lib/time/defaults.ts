import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

interface TimeRangeDropdownOptions {
  lastN: string[];
  toDate: string[];
  previousComplete: string[];
}

const lastNRanges = {
  minute: ["10m~", "15m~", "30m~", "60m~", "90m~", "120m~"],
  hour: ["6H~", "12H~", "24H~", "36H~", "48H~", "72H~"],
  day: ["3D~", "7D~", "14D~", "30D~", "90D~", "180D~"],
  week: ["2W~", "4W~", "6W~", "12W~", "26W~", "52W~"],
  month: ["3M~", "6M~", "9M~", "12M~", "24M~", "36M~"],
  quarter: ["2Q~", "3Q~", "4Q~", "6Q~", "8Q~", "12Q~"],
  year: ["2Y~", "3Y~", "5Y~", "7Y~", "10Y~", "15Y~"],
  mixed: ["6H~", "24H~", "7D~", "14D~", "4W~", "3M~", "12M~"],
};

export const timeRangeDefaultsByGrain: Record<
  V1TimeGrain,
  TimeRangeDropdownOptions
> = {
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: {
    lastN: lastNRanges.mixed,
    toDate: ["DTH~", "WTD~", "MTD~", "QTD~", "YTD~"],
    previousComplete: ["-1D", "-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: {
    lastN: lastNRanges.minute,
    toDate: ["DTH~", "WTD~", "MTD~", "QTD~", "YTD~"],
    previousComplete: ["Last 7 days", "Last 30 days", "Last 90 days"],
  },
  [V1TimeGrain.TIME_GRAIN_SECOND]: {
    lastN: lastNRanges.minute,
    toDate: ["mTs~", "HTs~", "DTs~", "WTs~", "MTs~", "QTs~", "YTs~"],
    previousComplete: ["-1D", "-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_MINUTE]: {
    lastN: lastNRanges.minute,
    toDate: ["HTm~", "DTm~", "WTm~", "MTm~", "QTm~", "YTm~"],
    previousComplete: ["-1D", "-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_HOUR]: {
    lastN: lastNRanges.hour,
    toDate: ["DTH~", "WTH~", "MTH~", "QTH~", "YTH~"],
    previousComplete: ["-1D", "-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_DAY]: {
    lastN: lastNRanges.day,
    toDate: ["WTD~", "MTD~", "QTD~", "YTD~"],
    previousComplete: ["-1D", "-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_WEEK]: {
    lastN: lastNRanges.week,
    toDate: ["MTW~", "QTW~", "YTW~"],
    previousComplete: ["-1W", "-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_MONTH]: {
    lastN: lastNRanges.month,
    toDate: ["QTM~", "YTM~"],
    previousComplete: ["-1M", "-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_QUARTER]: {
    lastN: lastNRanges.quarter,
    toDate: ["YTQ~"],
    previousComplete: ["-1Q", "-1Y"],
  },
  [V1TimeGrain.TIME_GRAIN_YEAR]: {
    lastN: lastNRanges.year,
    toDate: [],
    previousComplete: ["-1Y"],
  },
};
