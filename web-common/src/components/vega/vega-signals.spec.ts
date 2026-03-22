import { describe, it, expect } from "vitest";
import {
  resolveSignalField,
  resolveSignalTimeField,
  resolveSignalIntervalField,
} from "./vega-signals";

describe("resolveSignalField", () => {
  it("returns the first element of an array field", () => {
    expect(resolveSignalField({ dimension: ["US"] }, "dimension")).toBe("US");
  });

  it("returns undefined for non-array field", () => {
    expect(
      resolveSignalField({ dimension: "US" }, "dimension"),
    ).toBeUndefined();
  });

  it("returns undefined for missing field", () => {
    expect(resolveSignalField({ other: [1] }, "dimension")).toBeUndefined();
  });

  it("returns undefined for non-object value", () => {
    expect(resolveSignalField(null, "dimension")).toBeUndefined();
    expect(resolveSignalField(undefined, "dimension")).toBeUndefined();
  });
});

describe("resolveSignalTimeField", () => {
  const epoch = new Date("2024-01-15T00:00:00Z").getTime();

  // --- With temporalField hint (component chart path) ---

  it("finds time by temporalField hint with timeUnit prefix", () => {
    const signal = { yearmonthdate_timestamp: [epoch] };
    expect(resolveSignalTimeField(signal, "timestamp")).toEqual(
      new Date(epoch),
    );
  });

  it("finds time by temporalField hint with exact match (no timeUnit)", () => {
    const signal = { timestamp: [epoch] };
    expect(resolveSignalTimeField(signal, "timestamp")).toEqual(
      new Date(epoch),
    );
  });

  it("finds time for arbitrary field names like 'created_on'", () => {
    const signal = { yearmonthdate_created_on: [epoch] };
    expect(resolveSignalTimeField(signal, "created_on")).toEqual(
      new Date(epoch),
    );
  });

  // --- Without temporalField hint (_ts fallback, TDDAlternateChart path) ---

  it("falls back to _ts suffix when no hint provided (yearmonthdate_ts)", () => {
    const signal = { yearmonthdate_ts: [epoch] };
    expect(resolveSignalTimeField(signal)).toEqual(new Date(epoch));
  });

  it("falls back to _ts suffix for bare _ts key", () => {
    const signal = { some_ts: [epoch] };
    expect(resolveSignalTimeField(signal)).toEqual(new Date(epoch));
  });

  // --- Edge cases ---

  it("returns undefined when no keys match and no hint", () => {
    const signal = { yearmonthdate_timestamp: [epoch] };
    expect(resolveSignalTimeField(signal)).toBeUndefined();
  });

  it("returns undefined for empty object", () => {
    expect(resolveSignalTimeField({})).toBeUndefined();
  });

  it("returns undefined for null", () => {
    expect(resolveSignalTimeField(null)).toBeUndefined();
  });

  it("returns undefined when hint doesn't match any key", () => {
    const signal = { yearmonthdate_other: [epoch] };
    expect(resolveSignalTimeField(signal, "timestamp")).toBeUndefined();
  });
});

describe("resolveSignalIntervalField", () => {
  const start = new Date("2024-01-01T00:00:00Z").getTime();
  const end = new Date("2024-01-31T00:00:00Z").getTime();

  it("finds interval from 'ts' key", () => {
    const signal = { ts: [start, end] };
    const result = resolveSignalIntervalField(signal);
    expect(result).toEqual({ start: new Date(start), end: new Date(end) });
  });

  it("finds interval from key ending with '_ts'", () => {
    const signal = { yearmonthdate_ts: [start, end] };
    const result = resolveSignalIntervalField(signal);
    expect(result).toEqual({ start: new Date(start), end: new Date(end) });
  });

  it("falls back to any 2-element array key", () => {
    const signal = { yearmonthdate_timestamp: [start, end] };
    const result = resolveSignalIntervalField(signal);
    expect(result).toEqual({ start: new Date(start), end: new Date(end) });
  });

  it("returns undefined for empty object", () => {
    expect(resolveSignalIntervalField({})).toBeUndefined();
  });
});
