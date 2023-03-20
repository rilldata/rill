import {
  Period,
  RelativeTimeTransformation,
  TimeOffsetType,
  TimeTruncationType,
} from "../time-types";
import { transformDate } from "./";

function offsetOperation(
  duration: string,
  operationType: TimeOffsetType
): RelativeTimeTransformation {
  return {
    duration,
    operationType,
  };
}

function truncation(
  period: Period,
  truncationType: TimeTruncationType
): RelativeTimeTransformation {
  return {
    period,
    truncationType,
  };
}

const subtract = (duration) =>
  offsetOperation(duration, TimeOffsetType.SUBTRACT);
const add = (duration) => offsetOperation(duration, TimeOffsetType.ADD);
const startOf = (period) =>
  truncation(period, TimeTruncationType.START_OF_PERIOD);
const endOf = (period) => truncation(period, TimeTruncationType.END_OF_PERIOD);

const referenceTime = new Date(`2023-03-05T12:00:00+0000`);

const transformations = [
  {
    label: "should return the same date if no transformations are supplied",
    input: {
      referenceTime,
      transformation: [],
    },
    output: referenceTime,
  },
  {
    label: "should get this time last year",
    input: {
      referenceTime,
      transformation: [subtract("P1Y")],
    },
    output: new Date(`2022-03-05T12:00:00+0000`),
  },
  {
    label: "should subtract a day",
    input: {
      referenceTime,
      transformation: [subtract("P1D")],
    },
    output: new Date(`2023-03-04T12:00:00+0000`),
  },
  {
    label: "should add a day",
    input: {
      referenceTime,
      transformation: [add("P1D")],
    },
    output: new Date(`2023-03-06T12:00:00+0000`),
  },
  {
    label: "should subtract then add a day to get same time",
    input: {
      referenceTime,
      transformation: [subtract("P1D"), add("P1D")],
    },
    output: referenceTime,
  },
  {
    label: "should get beginning of month",
    input: {
      referenceTime,
      transformation: [startOf(Period.MONTH)],
    },
    output: new Date(`2023-03-01T00:00:00+0000`),
  },
  {
    label: "should get end of month",
    input: {
      referenceTime,
      transformation: [endOf(Period.MONTH)],
    },
    output: new Date(`2023-03-31T23:59:59.999+0000`),
  },

  {
    label: "should correctly land with start if we do start, end, start",
    input: {
      referenceTime,
      transformation: [
        startOf(Period.MONTH),
        endOf(Period.MONTH),
        startOf(Period.MONTH),
      ],
    },
    output: new Date(`2023-03-01T00:00:00+0000`),
  },
  {
    label: "should correctly land with end if we do end, start, end",
    input: {
      referenceTime,
      transformation: [
        endOf(Period.MONTH),
        startOf(Period.MONTH),
        endOf(Period.MONTH),
      ],
    },
    output: new Date(`2023-03-31T23:59:59.999+0000`),
  },
  {
    label:
      "should offset the reference time to the beginning of the previous hour",
    input: {
      referenceTime,
      transformation: [subtract("PT1H"), startOf(Period.HOUR)],
    },
    output: new Date(`2023-03-05T11:00:00+0000`),
  },
  {
    label: "should offset the reference time to the end of the previous hour",
    input: {
      referenceTime,
      transformation: [subtract("PT1H"), endOf(Period.HOUR)],
    },
    output: new Date(`2023-03-05T11:59:59.999+0000`),
  },
];

describe("transformDate", () => {
  for (const transformation of transformations) {
    it(transformation.label, () => {
      expect(
        transformDate(
          transformation.input.referenceTime,
          transformation.input.transformation
        ).toISOString()
      ).toEqual(transformation.output.toISOString());
    });
  }
});
