import { resolveChartTimeZone } from "@rilldata/web-common/features/chat/core/messages/chart/chart-block";
import type { V1ExploreSpec } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

describe("resolveChartTimeZone", () => {
  const shanghaiExplore: V1ExploreSpec = {
    timeZones: ["Asia/Shanghai", "UTC"],
  };

  it("keeps an explicit spec time_zone authoritative", () => {
    expect(resolveChartTimeZone("America/New_York", shanghaiExplore)).toBe(
      "America/New_York",
    );
  });

  it("inherits the explore default when the spec omits time_zone", () => {
    expect(resolveChartTimeZone(undefined, shanghaiExplore)).toBe(
      "Asia/Shanghai",
    );
  });

  it("inherits defaultPreset.timezone when that is the explore default", () => {
    const explore: V1ExploreSpec = {
      defaultPreset: { timezone: "Europe/Paris" },
      timeZones: ["UTC"],
    };
    expect(resolveChartTimeZone(undefined, explore)).toBe("Europe/Paris");
  });

  it("falls back to UTC when the explore lists no time zones", () => {
    expect(resolveChartTimeZone(undefined, {})).toBe("UTC");
  });

  it("falls back to UTC when explore context is absent", () => {
    expect(resolveChartTimeZone(undefined, undefined)).toBe("UTC");
  });

  it("does not treat an empty spec time_zone as explicit", () => {
    expect(resolveChartTimeZone("", shanghaiExplore)).toBe("Asia/Shanghai");
  });
});
