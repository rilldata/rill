import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  getFilterForComparedDimension,
  getTimeDimensionForComparison,
  niceMeasureExtents,
  transformAggregateDimensionData,
} from "./utils";

describe("niceMeasureExtents", () => {
  it("should return [0, 1] if both values are 0", () => {
    expect(niceMeasureExtents([0, 0], 1)).toEqual([0, 1]);
  });
  it("should ensure that if the minimum is positive, it returns 0", () => {
    expect(niceMeasureExtents([1, 2], 1)).toEqual([0, 2]);
  });
  it("should ensure that if the maximum is negative, it returns 0", () => {
    expect(niceMeasureExtents([-5, -2], 1)).toEqual([-5, 0]);
  });
  it("should inflate the minimum if it is negative", () => {
    expect(niceMeasureExtents([-5, 0], 2)).toEqual([-10, 0]);
  });
  it("should inflate the maximum if it is positive", () => {
    expect(niceMeasureExtents([0, 5], 2)).toEqual([0, 10]);
  });
});

describe("getFilterForComparedDimension", () => {
  it("should remove filters for selected dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("country", ["IN"], true),
    ]);

    const result = getFilterForComparedDimension(dimensionName, filters);

    expect(result).toEqual(
      createAndExpression([createInExpression("company", ["zoom"])]),
    );
  });

  it("should not modify filters for unrelated dimensions", () => {
    const dimensionName = "country";

    const filters = createAndExpression([
      createInExpression("company", ["zoom"]),
      createInExpression("device", ["mobile"], true),
    ]);

    const result = getFilterForComparedDimension(dimensionName, filters);

    expect(result).toEqual(filters);
  });
});

describe("getTimeDimensionForComparison", () => {
  it("requests the time dimension by name when comparing by another dimension", () => {
    expect(
      getTimeDimensionForComparison(
        "publisher",
        "ts",
        V1TimeGrain.TIME_GRAIN_MONTH,
        "UTC",
      ),
    ).toEqual({
      dimension: {
        name: "ts",
        timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
        timeZone: "UTC",
      },
      field: "ts",
    });
  });

  it("aliases the time floor when comparing by the time dimension itself", () => {
    expect(
      getTimeDimensionForComparison(
        "ts",
        "ts",
        V1TimeGrain.TIME_GRAIN_MONTH,
        "UTC",
      ),
    ).toEqual({
      dimension: {
        name: "ts",
        timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
        timeZone: "UTC",
        alias: "ts_rill_TIME_GRAIN_MONTH",
      },
      field: "ts_rill_TIME_GRAIN_MONTH",
    });
  });
});

describe("transformAggregateDimensionData", () => {
  it("fills the cells of raw timestamps that share a bucket when comparing by the time dimension", () => {
    const rows = transformAggregateDimensionData(
      "ts_rill_TIME_GRAIN_MONTH",
      "ts",
      ["events"],
      ["2024-01-05T00:00:00Z", "2024-01-20T00:00:00Z", "2024-02-03T00:00:00Z"],
      [{ ts: "2024-01-01T00:00:00Z" }, { ts: "2024-02-01T00:00:00Z" }],
      [
        {
          ts: "2024-01-05T00:00:00Z",
          ts_rill_TIME_GRAIN_MONTH: "2024-01-01T00:00:00Z",
          events: 1,
        },
        {
          ts: "2024-01-20T00:00:00Z",
          ts_rill_TIME_GRAIN_MONTH: "2024-01-01T00:00:00Z",
          events: 2,
        },
        {
          ts: "2024-02-03T00:00:00Z",
          ts_rill_TIME_GRAIN_MONTH: "2024-02-01T00:00:00Z",
          events: 4,
        },
      ],
    );

    expect(rows.map((row) => row.map((cell) => cell.records?.events))).toEqual([
      [1, null],
      [2, null],
      [null, 4],
    ]);
  });
});
