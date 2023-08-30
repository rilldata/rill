import { prepareTimeSeries } from "./utils";
import { adjustOffsetForZone } from "@rilldata/web-common/lib/convertTimestampPreview";
import { expect, test } from "vitest";

test("should fill in missing intervals", () => {
  const original = [
    { ts: "2023-04-14T11:00:00Z" },
    { ts: "2023-04-14T12:00:00Z" },
  ];

  const result = prepareTimeSeries(
    original,
    null,
    "PT1H",
    "UTC",
    "2023-04-14T10:00:00Z",
    "2023-04-14T13:00:00Z"
  );
  const ts = result.map((r) => formatDateToEnUS(r.ts));

  expect(ts[0]).toBe("Apr 14, 2023, 10:00 AM");
  expect(ts[1]).toBe("Apr 14, 2023, 11:00 AM");
  expect(ts[2]).toBe("Apr 14, 2023, 12:00 PM");
});

test("should fill in missing intervals in winter", () => {
  const original = [
    { ts: "2023-01-14T11:00:00Z" },
    { ts: "2023-01-14T12:00:00Z" },
  ];

  const result = prepareTimeSeries(
    original,
    null,
    "PT1H",
    "UTC",
    "2023-01-14T10:00:00Z",
    "2023-01-14T13:00:00Z"
  );
  const ts = result.map((r) => formatDateToEnUS(r.ts));

  expect(ts[0]).toBe("Jan 14, 2023, 10:00 AM");
  expect(ts[1]).toBe("Jan 14, 2023, 11:00 AM");
  expect(ts[2]).toBe("Jan 14, 2023, 12:00 PM");
});

test("should fill in missing intervals, CET", () => {
  const original = [
    { ts: "2023-01-14T11:00:00Z" },
    { ts: "2023-01-14T12:00:00Z" },
  ];

  const result = prepareTimeSeries(
    original,
    null,
    "PT1H",
    "CET",
    "2023-01-14T10:00:00Z",
    "2023-01-14T13:00:00Z"
  );
  const ts = result.map((r) => formatDateToEnUS(r.ts));

  expect(ts[0]).toBe("Jan 14, 2023, 11:00 AM");
  expect(ts[1]).toBe("Jan 14, 2023, 12:00 PM");
  expect(ts[2]).toBe("Jan 14, 2023, 1:00 PM");
});

function formatDateToEnUS(ts: Date): string {
  return ts.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "numeric",
  });
}

test("adjusts the timestamp for the given timezone", () => {
  let ts = new Date("2020-01-01T00:00:00Z");
  let zone = "America/New_York";

  let adjusted = adjustOffsetForZone(ts, zone);

  expect(
    adjusted.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    })
  ).toEqual("Dec 31, 2019, 7:00 PM");

  ts = new Date("2020-01-01T00:00:00Z");
  zone = "America/Los_Angeles";

  adjusted = adjustOffsetForZone(ts, zone);

  expect(
    adjusted.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    })
  ).toEqual("Dec 31, 2019, 4:00 PM");
});

test("comparison, should fill in missing intervals", () => {
  const original = [
    { ts: "2023-04-14T11:00:00Z" },
    { ts: "2023-04-14T12:00:00Z" },
  ];

  const comparison = [
    { ts: "2023-04-15T11:00:00Z" },
    { ts: "2023-04-15T12:00:00Z" },
  ];

  const result = prepareTimeSeries(
    original,
    comparison,
    "PT1H",
    "UTC",
    "2023-04-14T10:00:00Z",
    "2023-04-14T13:00:00Z",
    "2023-04-15T10:00:00Z",
    "2023-04-15T13:00:00Z"
  );

  const ts = result.map((r) => formatDateToEnUS(r.ts));

  const comp = result.map((r) => {
    return formatDateToEnUS(r["comparison.ts"]);
  });

  expect(ts[0]).toBe("Apr 14, 2023, 10:00 AM");
  expect(ts[1]).toBe("Apr 14, 2023, 11:00 AM");
  expect(ts[2]).toBe("Apr 14, 2023, 12:00 PM");

  expect(comp[0]).toBe("Apr 15, 2023, 10:00 AM");
  expect(comp[1]).toBe("Apr 15, 2023, 11:00 AM");
  expect(comp[2]).toBe("Apr 15, 2023, 12:00 PM");
});

test("comparison, should fill in missing intervals, America/Argentina/Buenos_Aires (no DST)", () => {
  const original = [
    { ts: "2023-04-14T10:00:00Z" },
    { ts: "2023-04-14T11:00:00Z" },
  ];

  const comparison = [
    { ts: "2023-04-15T10:00:00Z" },
    { ts: "2023-04-15T11:00:00Z" },
  ];

  const result = prepareTimeSeries(
    original,
    comparison,
    "PT1H",
    "America/Argentina/Buenos_Aires",
    "2023-04-14T10:00:00Z",
    "2023-04-14T13:00:00Z",
    "2023-04-15T10:00:00Z",
    "2023-04-15T13:00:00Z"
  );

  const ts = result.map((r) => formatDateToEnUS(r.ts));

  const comp = result.map((r) => {
    return formatDateToEnUS(r["comparison.ts"]);
  });

  expect(ts[0]).toBe("Apr 14, 2023, 7:00 AM");
  expect(ts[1]).toBe("Apr 14, 2023, 8:00 AM");
  expect(ts[2]).toBe("Apr 14, 2023, 9:00 AM");

  expect(comp[0]).toBe("Apr 15, 2023, 7:00 AM");
  expect(comp[1]).toBe("Apr 15, 2023, 8:00 AM");
  expect(comp[2]).toBe("Apr 15, 2023, 9:00 AM");
});

test("should include original records", () => {
  const original = [
    {
      ts: "2020-01-01T00:00:00Z",
      records: {
        clicks: 100,
        revenue: 10,
      },
    },
  ];

  const result = prepareTimeSeries(
    original,
    null,
    "PT1H",
    "UTC",
    "2020-01-01T00:00:00Z",
    "2020-01-01T01:00:00Z"
  );

  expect(result[0].clicks).toEqual(100);
  expect(result[0].revenue).toEqual(10);
});

test("should include comparison records", () => {
  const original = [
    {
      ts: "2020-01-01T00:00:00Z",
      records: {
        clicks: 100,
        revenue: 10,
      },
    },
  ];

  const comparison = [
    {
      ts: "2020-01-02T00:00:00Z",
      records: {
        clicks: 200,
        revenue: 20,
      },
    },
  ];

  const result = prepareTimeSeries(
    original,
    comparison,
    "PT1H",
    "UTC",
    "2020-01-01T00:00:00Z",
    "2020-01-01T01:00:00Z",
    "2020-01-02T00:00:00Z",
    "2020-01-02T01:00:00Z"
  );

  expect(result[0].clicks).toEqual(100);
  expect(result[0].revenue).toEqual(10);
  expect(result[0]["comparison.clicks"]).toEqual(200);
  expect(result[0]["comparison.revenue"]).toEqual(20);
});

test("should include comparison records", () => {
  const original = [
    {
      ts: "2020-01-01T00:00:00Z",
      records: {
        clicks: 100,
        revenue: 10,
      },
    },
  ];

  const comparison = [
    {
      ts: "2020-01-02T00:00:00Z",
      records: {
        clicks: 200,
        revenue: 20,
      },
    },
  ];

  const result = prepareTimeSeries(
    original,
    comparison,
    "PT1H",
    "UTC",
    "2020-01-01T00:00:00Z",
    "2020-01-01T01:00:00Z",
    "2020-01-02T00:00:00Z",
    "2020-01-02T01:00:00Z"
  );

  expect(result[0].clicks).toEqual(100);
  expect(result[0].revenue).toEqual(10);
  expect(result[0]["comparison.clicks"]).toEqual(200);
  expect(result[0]["comparison.revenue"]).toEqual(20);
});
