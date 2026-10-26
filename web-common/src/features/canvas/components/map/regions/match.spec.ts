import type * as GeoJSON from "geojson";
import { describe, expect, it } from "vitest";
import { calculateWrappedBounds } from "../map-utils";
import { buildRegionIndex, matchRegions, normalizeRegionValue } from "./match";

function region(
  properties: Record<string, unknown>,
  coordinates: number[][][] = [],
): GeoJSON.Feature {
  return {
    type: "Feature",
    properties,
    geometry: { type: "Polygon", coordinates },
  };
}

const COUNTRY_KEYS = ["iso_a2", "iso_a3", "iso_n3", "name"];

const countries = [
  region({
    id: "USA",
    name: "United States of America",
    iso_a2: "US",
    iso_a3: "USA",
    iso_n3: "840",
    aliases: "United States|USA|U.S.A.",
  }),
  region({
    id: "AUS",
    name: "Australia",
    iso_a2: "AU",
    iso_a3: "AUS",
    iso_n3: "036",
  }),
  region({ id: "IOA", name: "Indian Ocean Ter.", iso_a2: "AU" }),
  region({
    id: "CIV",
    name: "Côte d'Ivoire",
    iso_a2: "CI",
    iso_a3: "CIV",
    aliases: "Ivory Coast",
  }),
  region({ id: "SVK", name: "Slovakia", iso_a2: "SK", iso_a3: "SVK" }),
  // An alias that collides with another country's code must not steal it.
  region({ id: "KOR", name: "South Korea", iso_a2: "KR", aliases: "SK" }),
];

describe("normalizeRegionValue", () => {
  it("ignores case, accents, punctuation and whitespace", () => {
    expect(normalizeRegionValue("Côte d'Ivoire")).toBe("cotedivoire");
    expect(normalizeRegionValue("U.S.A.")).toBe("usa");
    expect(normalizeRegionValue("  New   York ")).toBe("newyork");
  });

  it("drops leading zeros from numeric codes", () => {
    expect(normalizeRegionValue("036")).toBe("36");
    expect(normalizeRegionValue(36)).toBe("36");
    expect(normalizeRegionValue("0")).toBe("0");
  });
});

describe("matchRegions", () => {
  const index = buildRegionIndex(countries, COUNTRY_KEYS);

  it("matches codes, names and aliases", () => {
    const rows = [
      { country: "us", revenue: 10 },
      { country: "Ivory Coast", revenue: 5 },
      { country: 36, revenue: 3 },
    ];
    const { values, unmatched } = matchRegions(
      rows,
      "country",
      "revenue",
      index,
    );
    expect(values.get("USA")).toEqual({ value: 10, label: "us" });
    expect(values.get("CIV")?.value).toBe(5);
    expect(values.get("AUS")?.value).toBe(3);
    expect(unmatched).toEqual([]);
  });

  it("matches every region that shares a code", () => {
    const { values } = matchRegions(
      [{ country: "AU", revenue: 1 }],
      "country",
      "revenue",
      index,
    );
    expect([...values.keys()].sort()).toEqual(["AUS", "IOA"]);
  });

  it("prefers the key listed first when keys collide", () => {
    const { values } = matchRegions(
      [{ country: "SK", revenue: 1 }],
      "country",
      "revenue",
      index,
    );
    expect([...values.keys()]).toEqual(["SVK"]);
  });

  it("keeps the first row when several rows match a region", () => {
    const { values } = matchRegions(
      [
        { country: "US", revenue: 1 },
        { country: "USA", revenue: 2 },
      ],
      "country",
      "revenue",
      index,
    );
    expect(values.get("USA")?.value).toBe(1);
  });

  it("reports distinct unmatched values and skips nulls", () => {
    const { unmatched } = matchRegions(
      [
        { country: "Atlantis", revenue: 1 },
        { country: "Atlantis", revenue: 2 },
        { country: null, revenue: 3 },
        { country: "US", revenue: null },
      ],
      "country",
      "revenue",
      index,
    );
    expect(unmatched).toEqual(["Atlantis"]);
  });

  it("only matches the selected key", () => {
    const isoOnly = buildRegionIndex(countries, ["iso_a3"]);
    const { values, unmatched } = matchRegions(
      [
        { country: "USA", revenue: 1 },
        { country: "US", revenue: 2 },
      ],
      "country",
      "revenue",
      isoOnly,
    );
    expect([...values.keys()]).toEqual(["USA"]);
    expect(unmatched).toEqual(["US"]);
  });
});

describe("calculateWrappedBounds", () => {
  it("fits shapes that cross the antimeridian without spanning the globe", () => {
    // Shaped like Russia: most of it in the east, the tip of Chukotka west of the antimeridian.
    const russia = region({ id: "RUS" }, [
      [
        [30, 50],
        [180, 70],
        [30, 70],
        [30, 50],
      ],
      [
        [-180, 65],
        [-170, 66],
        [-180, 66],
        [-180, 65],
      ],
    ]);
    expect(calculateWrappedBounds([russia])).toEqual([
      [30, 50],
      [190, 70],
    ]);
  });

  it("keeps regular bounds for shapes in one hemisphere", () => {
    const shape = region({ id: "X" }, [
      [
        [-120, 30],
        [-70, 45],
        [-120, 45],
        [-120, 30],
      ],
    ]);
    expect(calculateWrappedBounds([shape])).toEqual([
      [-120, 30],
      [-70, 45],
    ]);
  });

  it("returns null without coordinates", () => {
    expect(calculateWrappedBounds([])).toBeNull();
  });
});
