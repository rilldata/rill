import {
  Period,
  RelativeTimeTransformation,
  TimeOffsetType,
  TimeTruncationType,
} from "../types";
import { getDurationMultiple, transformDate } from "./";

import { durationToMillis } from "../grains";
import { getEndOfPeriod, getOffset, getStartOfPeriod, getTimeWidth } from "./";

describe("getStartOfPeriod", () => {
  it("should return the start of the week for given date", () => {
    const timeGrain = getStartOfPeriod(new Date("2020-03-15"), Period.WEEK);
    expect(timeGrain).toEqual(new Date("2020-03-09"));
  });
  it("should return the start of month for given date", () => {
    const timeGrain = getStartOfPeriod(new Date("2020-03-15"), Period.MONTH);
    expect(timeGrain).toEqual(new Date("2020-03-01"));
  });
});

describe("getEndOfPeriod", () => {
  it("should return the end of the week for given date", () => {
    const timeGrain = getEndOfPeriod(new Date("2020-03-15"), Period.WEEK);
    expect(timeGrain).toEqual(new Date("2020-03-15T23:59:59.999Z"));
  });
  it("should return the end of month for given date", () => {
    const timeGrain = getEndOfPeriod(new Date("2020-02-15"), Period.MONTH);
    // leap year!
    expect(timeGrain).toEqual(new Date("2020-02-29T23:59:59.999Z"));
  });
});

describe("getOffset", () => {
  it("should add correct amount of time for given date", () => {
    const timeGrain = getOffset(
      new Date("2020-02-15"),
      "P2W",
      TimeOffsetType.ADD
    );
    expect(timeGrain).toEqual(new Date("2020-02-29"));
  });
  it("should subtract correct amount of time for given date", () => {
    const timeGrain = getOffset(
      new Date("2020-02-15"),
      "P2M",
      TimeOffsetType.SUBTRACT
    );
    expect(timeGrain).toEqual(new Date("2019-12-15"));
  });
});

describe("getTimeWidth", () => {
  it("should give correct amount of time width in milliseconds for given dates", () => {
    const timeGrain = getTimeWidth(
      new Date("2020-03-15"),
      new Date("2020-04-01")
    );
    expect(timeGrain).toEqual(durationToMillis("P1D") * 17);
  });
});

/** core transformation tests. */

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

describe("getDurationMultiple", () => {
  it("should return the half of given week duration", () => {
    const duration = getDurationMultiple("P1W", 0.5);
    expect(duration).toEqual("P3DT12H");
  });
  it("should return the triple of given arbitary duration", () => {
    const duration = getDurationMultiple("P2DT12H", 3);
    expect(duration).toEqual("P7DT12H");
  });
});
