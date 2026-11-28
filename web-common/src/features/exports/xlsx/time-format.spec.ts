import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { WorkbookColumnType } from "@rilldata/web-common/proto/gen/rill/runtime/v1/workbook_pb";
import { describe, expect, it } from "vitest";
import {
  excelDateFormat,
  formatTimeLabel,
  localIsoInZone,
  timeColumnType,
} from "./time-format";

describe("localIsoInZone", () => {
  it("writes an ISO string as the wall-clock time in the zone", () => {
    expect(localIsoInZone("2026-01-04T00:00:00Z", "America/New_York")).toBe(
      "2026-01-03T19:00:00",
    );
    expect(localIsoInZone("2026-01-04T00:00:00Z", "UTC")).toBe(
      "2026-01-04T00:00:00",
    );
  });

  it("keeps the local fields of a date the chart pipeline already shifted", () => {
    const date = new Date(2026, 0, 4, 9, 30);
    expect(localIsoInZone(date, "Asia/Kolkata")).toBe("2026-01-04T09:30:00");
  });

  it("passes other values through", () => {
    expect(localIsoInZone("not a date", "UTC")).toBe("not a date");
    expect(localIsoInZone(null, "UTC")).toBeNull();
  });
});

describe("grain formats", () => {
  it("types day and coarser grains as dates", () => {
    expect(timeColumnType(V1TimeGrain.TIME_GRAIN_DAY)).toBe(
      WorkbookColumnType.DATE,
    );
    expect(timeColumnType(V1TimeGrain.TIME_GRAIN_MONTH)).toBe(
      WorkbookColumnType.DATE,
    );
    expect(timeColumnType(V1TimeGrain.TIME_GRAIN_HOUR)).toBe(
      WorkbookColumnType.TIMESTAMP,
    );
    expect(timeColumnType(undefined)).toBe(WorkbookColumnType.TIMESTAMP);
  });

  it("shows each grain at its own precision", () => {
    expect(excelDateFormat(V1TimeGrain.TIME_GRAIN_YEAR)).toBe("yyyy");
    expect(excelDateFormat(V1TimeGrain.TIME_GRAIN_MONTH)).toBe("mmm yyyy");
    expect(excelDateFormat(V1TimeGrain.TIME_GRAIN_WEEK)).toBe("yyyy-mm-dd");
    expect(excelDateFormat(V1TimeGrain.TIME_GRAIN_HOUR)).toBe(
      "yyyy-mm-dd hh:mm",
    );
    expect(excelDateFormat(V1TimeGrain.TIME_GRAIN_SECOND)).toBe(
      "yyyy-mm-dd hh:mm:ss",
    );
  });

  it("labels nested time values", () => {
    expect(
      formatTimeLabel(
        "2026-04-01T00:00:00Z",
        V1TimeGrain.TIME_GRAIN_QUARTER,
        "UTC",
      ),
    ).toBe("Q2 2026");
    expect(
      formatTimeLabel(
        "2026-04-01T00:00:00Z",
        V1TimeGrain.TIME_GRAIN_MONTH,
        "UTC",
      ),
    ).toBe("Apr 2026");
    expect(
      formatTimeLabel(
        "2026-04-01T13:00:00Z",
        V1TimeGrain.TIME_GRAIN_HOUR,
        "UTC",
      ),
    ).toBe("2026-04-01 13:00");
    expect(formatTimeLabel("nope", V1TimeGrain.TIME_GRAIN_DAY, "UTC")).toBe(
      "nope",
    );
  });
});
