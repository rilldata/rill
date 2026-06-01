import { buildTagIndex } from "@rilldata/web-common/components/menu/tag-utils";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  dimensionToChipData,
  measureToChipData,
  splitTagItems,
} from "./pivot-utils";
import { PivotChipType } from "./types";

const dim = (
  name: string,
  tags: string[] = [],
  displayName?: string,
): MetricsViewSpecDimension => ({
  name,
  displayName,
  tags,
});

const meas = (
  name: string,
  tags: string[] = [],
  displayName?: string,
): MetricsViewSpecMeasure => ({
  name,
  displayName,
  tags,
});

describe("dimensionToChipData", () => {
  it("uses displayName when present", () => {
    expect(dimensionToChipData(dim("country", [], "Country"))).toEqual({
      id: "country",
      title: "Country",
      type: PivotChipType.Dimension,
      description: undefined,
    });
  });

  it("falls back to name when displayName is missing", () => {
    expect(dimensionToChipData(dim("region"))).toEqual({
      id: "region",
      title: "region",
      type: PivotChipType.Dimension,
      description: undefined,
    });
  });
});

describe("measureToChipData", () => {
  it("projects a measure spec", () => {
    expect(measureToChipData(meas("revenue", [], "Revenue"))).toEqual({
      id: "revenue",
      title: "Revenue",
      type: PivotChipType.Measure,
      description: undefined,
    });
  });
});

describe("splitTagItems", () => {
  const dimensions = [
    dim("country", ["Geography", "Customer"], "Country"),
    dim("region", ["Geography"], "Region"),
    dim("segment", ["Customer"], "Segment"),
  ];
  const measures = [
    meas("revenue", ["Geography", "Finance"], "Revenue"),
    meas("profit", ["Finance"], "Profit"),
  ];
  const dimIndex = buildTagIndex(dimensions);
  const measIndex = buildTagIndex(measures);

  it("splits a mixed tag into dim chips and measure chips", () => {
    const { dimensions: dims, measures: meas } = splitTagItems(
      "Geography",
      dimIndex,
      measIndex,
    );
    expect(dims.map((c) => c.id)).toEqual(["country", "region"]);
    expect(meas.map((c) => c.id)).toEqual(["revenue"]);
    expect(dims.every((c) => c.type === PivotChipType.Dimension)).toBe(true);
    expect(meas.every((c) => c.type === PivotChipType.Measure)).toBe(true);
  });

  it("returns only dimensions for a pure-dimension tag", () => {
    const { dimensions: dims, measures: meas } = splitTagItems(
      "Customer",
      dimIndex,
      measIndex,
    );
    expect(dims.map((c) => c.id)).toEqual(["country", "segment"]);
    expect(meas).toEqual([]);
  });

  it("returns only measures for a pure-measure tag", () => {
    const { dimensions: dims, measures: meas } = splitTagItems(
      "Finance",
      dimIndex,
      measIndex,
    );
    expect(dims).toEqual([]);
    expect(meas.map((c) => c.id)).toEqual(["revenue", "profit"]);
  });

  it("returns empty arrays for an unknown tag", () => {
    expect(splitTagItems("Nope", dimIndex, measIndex)).toEqual({
      dimensions: [],
      measures: [],
    });
  });

  it("preserves spec order in output", () => {
    const orderedDims = [
      dim("a", ["T"]),
      dim("b", ["T"]),
      dim("c", ["T"]),
    ];
    const orderedMeas = [
      meas("x", ["T"]),
      meas("y", ["T"]),
    ];
    const { dimensions: dims, measures: meas2 } = splitTagItems(
      "T",
      buildTagIndex(orderedDims),
      buildTagIndex(orderedMeas),
    );
    expect(dims.map((c) => c.id)).toEqual(["a", "b", "c"]);
    expect(meas2.map((c) => c.id)).toEqual(["x", "y"]);
  });
});
