/**
 * This module defines configured presets for time ranges & time grains.
 * We define them as JSON objects primarily users will eventually be able to
 * manually define these in the dashboard configuration.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  AvailableTimeGrain,
  Period,
  RangePresetType,
  ReferencePoint,
  TimeGrain,
  TimeOffsetType,
  TimeRangeMeta,
  TimeTruncationType,
} from "./types";

/**
 * The "latest" window time ranges are defined as a set of time ranges that are
 * anchored to the latest data point in the dataset with a conceptually-fixed
 * lookback window. For example, the "Last 6 Hours" time range is anchored to
 * the latest data point in the dataset, and then looks back 6 hours from that
 * point.
 *
 * This description is not 100% accurate, of course, since the latest data point
 * may be during an incomplete period. For now, we are truncating to a reasonable
 * periodicity (e.g. to the start of the hour) and then applying the offset.
 */
export const LATEST_WINDOW_TIME_RANGES: Record<string, TimeRangeMeta> = {
  LAST_SIX_HOURS: {
    label: "Last 6 Hours",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        // start during the last full hour.
        { duration: "PT1H", operationType: TimeOffsetType.SUBTRACT },
        {
          period: Period.HOUR, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
        // then offset that by 6 hours
        { duration: "PT6H", operationType: TimeOffsetType.SUBTRACT }, // operation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },

  LAST_24_HOURS: {
    label: "Last 24 Hours",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1D", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.HOUR, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },

  LAST_7_DAYS: {
    label: "Last 7 Days",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1W", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.DAY, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  LAST_4_WEEKS: {
    label: "Last 4 Weeks",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P4W", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.DAY, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  LAST_YEAR: {
    label: "Last Year",
    rangePreset: RangePresetType.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "P1Y", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.DAY, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
};

/**
 * The "period to date" time ranges are defined as a set of time ranges that are
 * anchored to the latest data point in the dataset, with a start datetime
 * that's anchored to the beginning of the period that the latest data point is in.
 * For example, the "Today" time range is anchored to the latest data point in
 * the dataset, and then looks back to the start of that day.
 *
 * Like the latest window ranges, wetruncate the latest data point datetime to the
 * start of a reasonable period for now.
 */
export const PERIOD_TO_DATE_RANGES: Record<string, TimeRangeMeta> = {
  TODAY: {
    label: "Today",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  WEEK_TO_DATE: {
    label: "Week to Date",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.WEEK, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  MONTH_TO_DATE: {
    label: "Month to Date",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.MONTH, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  YEAR_TO_DATE: {
    label: "Year to Date",
    rangePreset: RangePresetType.PERIOD_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.YEAR, //TODO: How to handle user selected timegrains?
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        {
          period: Period.DAY,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
};

export const ALL_TIME = {
  label: "All Time",
  rangePreset: RangePresetType.ALL_TIME,
};

export const DEFAULT_TIME_RANGES: Record<string, TimeRangeMeta> = {
  ...LATEST_WINDOW_TIME_RANGES,
  ...PERIOD_TO_DATE_RANGES,
  ALL_TIME,
};

// This is a temporary fix for the default time range setting.
// We need to deprecate this once we have moved the default_time_range setting to operate
// on preset strings rather than ISO durations.
// See https://github.com/rilldata/rill-developer/issues/1961
export const TEMPORARY_DEFAULT_RANGE_TO_DURATIONS = {
  LAST_SIX_HOURS: "PT6H",
  LAST_24_HOURS: "P1D",
  LAST_7_DAYS: "P7D",
  LAST_4_WEEKS: "P4W",
  LAST_YEAR: "P1Y",
  TODAY: "P1D",
};

/**
 * A time grain is a unit of time that is used to group data points,
 * e.g. "hour" or "day". The time grain is used to aggregate records
 * for the purposes of time series visualization and analysis.
 */
export const TIME_GRAIN: Record<AvailableTimeGrain, TimeGrain> = {
  TIME_GRAIN_MINUTE: {
    grain: V1TimeGrain.TIME_GRAIN_MINUTE,
    label: "minute",
    duration: Period.MINUTE,
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    },
  },
  TIME_GRAIN_HOUR: {
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
    label: "hour",
    duration: Period.HOUR,
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
    },
  },
  TIME_GRAIN_DAY: {
    grain: V1TimeGrain.TIME_GRAIN_DAY,
    label: "day",
    duration: Period.DAY,
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_WEEK: {
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
    label: "week",
    duration: Period.WEEK,
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_MONTH: {
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
    label: "month",
    duration: Period.MONTH,
    formatDate: {
      year: "numeric",
      month: "short",
    },
  },
  TIME_GRAIN_YEAR: {
    grain: V1TimeGrain.TIME_GRAIN_YEAR,
    label: "year",
    duration: Period.YEAR,
    formatDate: {
      year: "numeric",
    },
  },
};
