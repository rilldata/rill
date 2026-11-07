import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { scaleLinear } from "d3-scale";
import { DateTime } from "luxon";
import { describe, expect, it } from "vitest";
import {
  type Annotation,
  AnnotationWidth,
  groupAnnotations,
} from "./annotation-utils";
import type { ChartConfig, ChartScales, TimeSeriesPoint } from "./types";

// Four daily data points from Feb 1 to Feb 4, 2024, plotted from x=0 to x=300.
const data: TimeSeriesPoint[] = [1, 2, 3, 4].map((day) => ({
  ts: DateTime.utc(2024, 2, day),
  value: day,
}));

const plotBounds = {
  left: 0,
  right: 300,
  top: 0,
  bottom: 100,
  width: 300,
  height: 100,
};
const config: ChartConfig = {
  width: 300,
  height: 100,
  margin: { top: 0, right: 0, bottom: 0, left: 0 },
  plotBounds,
};
const scales: ChartScales = {
  x: scaleLinear()
    .domain([0, data.length - 1])
    .range([0, 300]),
  y: scaleLinear().domain([0, 4]).range([100, 0]),
};

function annotation(time: string, timeEnd?: string): Annotation {
  return {
    time,
    timeEnd,
    description: time,
    startTime: DateTime.fromISO(time, { zone: "UTC" }),
    endTime: timeEnd ? DateTime.fromISO(timeEnd, { zone: "UTC" }) : undefined,
    formattedTimeOrRange: "",
  };
}

function group(annotations: Annotation[]) {
  return groupAnnotations(
    annotations,
    scales,
    data,
    config,
    V1TimeGrain.TIME_GRAIN_DAY,
  );
}

describe("groupAnnotations", () => {
  it("positions a point annotation at its grain bucket", () => {
    const groups = group([annotation("2024-02-03T06:00:00Z")]);

    expect(groups).toHaveLength(1);
    expect(groups[0].index).toBe(2);
    expect(groups[0].left).toBe(200);
    expect(groups[0].right).toBe(200 + AnnotationWidth);
    expect(groups[0].hasRange).toBe(false);
  });

  it("groups annotations that fall in the same bucket", () => {
    const groups = group([
      annotation("2024-02-02T01:00:00Z"),
      annotation("2024-02-02T23:00:00Z"),
      annotation("2024-02-04T00:00:00Z"),
    ]);

    expect(groups.map((g) => g.index)).toEqual([1, 3]);
    expect(groups[0].items).toHaveLength(2);
  });

  it("spans a range annotation from its start to its end bucket", () => {
    const groups = group([
      annotation("2024-02-02T00:00:00Z", "2024-02-03T00:00:00Z"),
    ]);

    expect(groups).toHaveLength(1);
    expect(groups[0].left).toBe(100);
    expect(groups[0].right).toBe(200);
    expect(groups[0].hasRange).toBe(true);
  });

  it("anchors a range that starts before the data to the first data point", () => {
    const groups = group([
      annotation("2024-01-30T00:00:00Z", "2024-02-03T00:00:00Z"),
    ]);

    expect(groups).toHaveLength(1);
    expect(groups[0].index).toBe(0);
    expect(groups[0].left).toBe(0);
    expect(groups[0].right).toBe(200);
    expect(groups[0].hasRange).toBe(true);
  });

  it("drops a range that ends exactly where the data starts", () => {
    const groups = group([
      annotation("2024-01-30T00:00:00Z", "2024-02-01T00:00:00Z"),
    ]);

    expect(groups).toEqual([]);
  });

  it("extends a range that ends after the data to the last data point", () => {
    const groups = group([
      annotation("2024-02-03T00:00:00Z", "2024-02-10T00:00:00Z"),
    ]);

    expect(groups).toHaveLength(1);
    expect(groups[0].left).toBe(200);
    expect(groups[0].right).toBe(300);
  });

  it("drops annotations that do not overlap the data", () => {
    const groups = group([
      annotation("2024-01-30T00:00:00Z"),
      annotation("2024-01-20T00:00:00Z", "2024-01-25T00:00:00Z"),
      annotation("2024-02-05T00:00:00Z"),
    ]);

    expect(groups).toEqual([]);
  });
});
