/**
 * This module defines configured presets for time ranges & time grains.
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

export const DEFAULT_TIME_RANGE_PRESETS = Object.keys(
  DEFAULT_TIME_RANGES
).reduce((acc, key) => {
  acc[key] = key;
  return acc;
}, {} as Record<string, string>);

/** Configuration for the admissible set of time grains. See the TimeGrain interface for more information about values. */
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
