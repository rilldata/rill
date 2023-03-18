import {
  Period,
  RangePreset,
  ReferencePoint,
  TimeOffsetType,
  TimeRangeMeta,
  TimeTruncationType,
} from "./time-types";

export const NOW_ANCHORED_TIMES: Record<string, TimeRangeMeta> = {
  LAST_SIX_HOURS: {
    label: "Last 6 Hours",
    rangePreset: RangePreset.OFFSET_ANCHORED,
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
    rangePreset: RangePreset.OFFSET_ANCHORED,
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
    rangePreset: RangePreset.OFFSET_ANCHORED,
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
    rangePreset: RangePreset.OFFSET_ANCHORED,
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
    rangePreset: RangePreset.OFFSET_ANCHORED,
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

export const PERIOD_ANCHORED_RANGES: Record<string, TimeRangeMeta> = {
  TODAY: {
    label: "Today",
    rangePreset: RangePreset.PERIOD_ANCHORED,
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
    rangePreset: RangePreset.PERIOD_ANCHORED,
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
    rangePreset: RangePreset.PERIOD_ANCHORED,
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
    rangePreset: RangePreset.PERIOD_ANCHORED,
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

export const DEFAULT_TIME_RANGES: Record<string, TimeRangeMeta> = {
  ...NOW_ANCHORED_TIMES,
  ...PERIOD_ANCHORED_RANGES,
  ALL_TIME: {
    label: "All Time",
    rangePreset: RangePreset.ALL_TIME,
  },
};
