import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { mergeComparisonRowsAcrossDST } from "./util";

describe("mergeComparisonRowsAcrossDST", () => {
  const tz = "America/Los_Angeles";

  it("merges base and comparison rows split by a DST transition", () => {
    const data = [
      { timestamp: "2026-11-06T07:00:00Z", m: null, m_prev: 391 },
      { timestamp: "2026-11-06T08:00:00Z", m: 413, m_prev: null },
      { timestamp: "2026-11-07T07:00:00Z", m: null, m_prev: 400 },
      { timestamp: "2026-11-07T08:00:00Z", m: 420, m_prev: null },
      { timestamp: "2026-11-09T08:00:00Z", m: 430, m_prev: 410 },
    ];
    expect(
      mergeComparisonRowsAcrossDST(
        data,
        ["timestamp"],
        [],
        tz,
        V1TimeGrain.TIME_GRAIN_DAY,
      ),
    ).toEqual([
      { timestamp: "2026-11-06T08:00:00Z", m: 413, m_prev: 391 },
      { timestamp: "2026-11-07T08:00:00Z", m: 420, m_prev: 400 },
      { timestamp: "2026-11-09T08:00:00Z", m: 430, m_prev: 410 },
    ]);
  });

  it("merges rows split by the spring-forward transition", () => {
    // Base days are PDT (07:00Z); comparison days are PST (08:00Z) shifted by 7 days.
    const data = [
      { timestamp: "2026-03-09T07:00:00Z", m: 413, m_prev: null },
      { timestamp: "2026-03-09T08:00:00Z", m: null, m_prev: 391 },
      { timestamp: "2026-03-10T07:00:00Z", m: 420, m_prev: null },
      { timestamp: "2026-03-10T08:00:00Z", m: null, m_prev: 400 },
    ];
    expect(
      mergeComparisonRowsAcrossDST(
        data,
        ["timestamp"],
        [],
        tz,
        V1TimeGrain.TIME_GRAIN_DAY,
      ),
    ).toEqual([
      { timestamp: "2026-03-09T07:00:00Z", m: 413, m_prev: 391 },
      { timestamp: "2026-03-10T07:00:00Z", m: 420, m_prev: 400 },
    ]);
  });

  it("merges weekly buckets split by a DST transition", () => {
    const data = [
      { timestamp: "2026-11-09T07:00:00Z", m: null, m_prev: 5 },
      { timestamp: "2026-11-09T08:00:00Z", m: 6, m_prev: null },
    ];
    expect(
      mergeComparisonRowsAcrossDST(
        data,
        ["timestamp"],
        [],
        tz,
        V1TimeGrain.TIME_GRAIN_WEEK,
      ),
    ).toEqual([{ timestamp: "2026-11-09T08:00:00Z", m: 6, m_prev: 5 }]);
  });

  it("keeps rows with different dimension values apart", () => {
    const data = [
      { timestamp: "2026-11-06T07:00:00Z", d: "a", m: null, m_prev: 1 },
      { timestamp: "2026-11-06T07:00:00Z", d: "b", m: null, m_prev: 2 },
      { timestamp: "2026-11-06T08:00:00Z", d: "a", m: 3, m_prev: null },
      { timestamp: "2026-11-06T08:00:00Z", d: "b", m: 4, m_prev: null },
    ];
    expect(
      mergeComparisonRowsAcrossDST(
        data,
        ["timestamp"],
        ["d"],
        tz,
        V1TimeGrain.TIME_GRAIN_DAY,
      ),
    ).toEqual([
      { timestamp: "2026-11-06T08:00:00Z", d: "a", m: 3, m_prev: 1 },
      { timestamp: "2026-11-06T08:00:00Z", d: "b", m: 4, m_prev: 2 },
    ]);
  });

  it("leaves sub-day grains untouched", () => {
    const data = [{ timestamp: "2026-11-06T07:00:00Z", m: 1 }];
    expect(
      mergeComparisonRowsAcrossDST(
        data,
        ["timestamp"],
        [],
        tz,
        V1TimeGrain.TIME_GRAIN_HOUR,
      ),
    ).toBe(data);
  });
});
