import { describe, expect, it } from "vitest";
import bar from "../../harness/fixtures/bar_chart.json";
import funnel from "../../harness/fixtures/funnel_chart.json";
import horizontal from "../../harness/fixtures/horizontal_bar.json";
import line from "../../harness/fixtures/line_chart.json";
import {
  availableGrains,
  controlOptions,
  describeView,
  patchSpec,
  previousPeriod,
  topNChoices,
  topNTarget,
} from "./controls";
import type { RillChartResult } from "./contract";

const result = (fixture: { result: { structuredContent?: unknown } }) =>
  fixture.result.structuredContent as RillChartResult;

const week = { start: "2023-09-09T00:00:00Z", end: "2023-09-16T00:00:00Z" };

describe("availableGrains", () => {
  it("offers grains that split the range into a readable number of buckets", () => {
    expect(
      availableGrains("line_chart", result(line).spec, week, "TIME_GRAIN_HOUR"),
    ).toEqual(["TIME_GRAIN_HOUR", "TIME_GRAIN_DAY"]);
    expect(
      availableGrains("line_chart", result(line).spec, week, "TIME_GRAIN_DAY"),
    ).toEqual(["TIME_GRAIN_DAY"]);
  });

  it("returns nothing without a temporal x axis", () => {
    expect(availableGrains("bar_chart", result(bar).spec, week)).toEqual([]);
  });
});

describe("topNTarget", () => {
  it("targets the category channel, also for horizontal bars", () => {
    expect(topNTarget("bar_chart", result(bar).spec)).toEqual({
      channel: "x",
      limit: 10,
    });
    expect(topNTarget("bar_chart", result(horizontal).spec)).toEqual({
      channel: "y",
      limit: 10,
    });
    expect(topNTarget("funnel_chart", result(funnel).spec)).toEqual({
      channel: "stage",
      limit: 15,
    });
    expect(topNTarget("line_chart", result(line).spec)).toBeUndefined();
  });
});

describe("topNChoices", () => {
  it("offers the presets plus the current limit when it is not one", () => {
    expect(topNChoices(undefined)).toEqual([5, 10, 20]);
    expect(topNChoices(10)).toEqual([5, 10, 20]);
    expect(topNChoices(15)).toEqual([5, 10, 15, 20]);
    expect(topNChoices(40)).toEqual([5, 10, 20, 40]);
  });
});

describe("patchSpec", () => {
  it("applies grain, top-N and comparison without mutating the input", () => {
    const spec = result(line).spec;
    const patched = patchSpec(
      spec,
      "line_chart",
      { grain: "TIME_GRAIN_HOUR", comparison: true },
      week,
    );
    expect(patched.time_grain).toBe("TIME_GRAIN_HOUR");
    expect(patched.comparison_time_range).toEqual({
      start: "2023-09-02T00:00:00.000Z",
      end: "2023-09-09T00:00:00Z",
    });
    expect(spec.time_grain).toBe("TIME_GRAIN_DAY");
    expect(spec.comparison_time_range).toBeUndefined();

    const limited = patchSpec(
      result(horizontal).spec,
      "bar_chart",
      { topN: 5 },
      week,
    );
    expect((limited.y as { limit: number }).limit).toBe(5);

    const off = patchSpec(patched, "line_chart", { comparison: false }, week);
    expect(off.comparison_time_range).toBeUndefined();
  });
});

describe("controlOptions", () => {
  it("summarizes what a chart can control", () => {
    expect(
      controlOptions("bar_chart", result(bar).spec, week, "TIME_GRAIN_HOUR"),
    ).toEqual({
      grains: [],
      topN: true,
      comparison: true,
    });
  });
});

describe("previousPeriod", () => {
  it("shifts the range back by its own length", () => {
    expect(previousPeriod(week)).toEqual({
      start: "2023-09-02T00:00:00.000Z",
      end: "2023-09-09T00:00:00Z",
    });
  });
});

describe("describeView", () => {
  it("states every control value for the model", () => {
    expect(
      describeView(
        "Total Bids over Time",
        { grain: "TIME_GRAIN_DAY", topN: 5, comparison: true },
        "Sep 9 – Sep 15, 2023",
      ),
    ).toBe(
      "Chart now shows Total Bids over Time: time grain day, top 5 values, comparison on, Sep 9 – Sep 15, 2023.",
    );
  });

  it("states the limit in effect even when it is not a preset", () => {
    expect(
      describeView("Total Bids funnel by Campaign Stage", { topN: 15 }),
    ).toBe(
      "Chart now shows Total Bids funnel by Campaign Stage: time grain n/a, top 15 values, comparison off.",
    );
  });

  it("marks the controls a chart does not offer", () => {
    expect(describeView("Total Bids split by Device", {})).toBe(
      "Chart now shows Total Bids split by Device: time grain n/a, top values n/a, comparison off.",
    );
  });
});
