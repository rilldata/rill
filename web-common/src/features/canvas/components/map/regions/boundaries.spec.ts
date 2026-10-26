import type * as GeoJSON from "geojson";
import { describe, expect, it, vi } from "vitest";
import {
  buildCustomFeatures,
  decodeBoundaries,
  isValidRegionBoundaries,
  regionIndexKeys,
} from "./boundaries";
import { buildRegionIndex } from "./match";

const square: GeoJSON.Polygon = {
  type: "Polygon",
  coordinates: [
    [
      [0, 0],
      [1, 0],
      [1, 1],
      [0, 1],
      [0, 0],
    ],
  ],
};

function feature(
  properties: Record<string, unknown> | null,
  geometry: GeoJSON.Geometry = square,
): GeoJSON.Feature {
  return { type: "Feature", properties, geometry };
}

function collection(...features: GeoJSON.Feature[]) {
  return { type: "FeatureCollection", features } as GeoJSON.FeatureCollection;
}

describe("decodeBoundaries", () => {
  it("returns a GeoJSON FeatureCollection as is", () => {
    const fc = collection(feature({ code: "501" }));
    expect(decodeBoundaries(fc)).toBe(fc);
  });

  it("decodes a TopoJSON topology with one object", () => {
    const topology = {
      type: "Topology",
      arcs: [
        [
          [0, 0],
          [1, 0],
          [1, 1],
          [0, 1],
          [0, 0],
        ],
      ],
      objects: {
        dma: {
          type: "GeometryCollection",
          geometries: [
            { type: "Polygon", arcs: [[0]], properties: { code: "501" } },
          ],
        },
      },
    };
    const fc = decodeBoundaries(topology);
    expect(fc.type).toBe("FeatureCollection");
    expect(fc.features).toHaveLength(1);
    expect(fc.features[0].properties).toEqual({ code: "501" });
    expect(fc.features[0].geometry.type).toBe("Polygon");
  });

  it("rejects a topology with several objects", () => {
    const topology = {
      type: "Topology",
      arcs: [],
      objects: {
        a: { type: "GeometryCollection", geometries: [] },
        b: { type: "GeometryCollection", geometries: [] },
      },
    };
    expect(() => decodeBoundaries(topology)).toThrow(/one object, found 2/);
  });

  it("rejects other JSON", () => {
    expect(() => decodeBoundaries({ type: "Feature" })).toThrow(
      /FeatureCollection or a TopoJSON Topology/,
    );
    expect(() => decodeBoundaries(null)).toThrow();
  });
});

describe("buildCustomFeatures", () => {
  it("sets ids and names from the configured properties", () => {
    const fc = buildCustomFeatures(
      collection(
        feature({ code: 501, label: "New York" }),
        feature({ code: "803" }),
      ),
      "code",
      "label",
    );
    expect(fc.features.map((f) => f.properties)).toEqual([
      { id: "501", name: "New York" },
      { id: "803", name: "803" },
    ]);
  });

  it("falls back to the feature id when the property is missing", () => {
    const fc = buildCustomFeatures(
      collection(
        { ...feature({ NAME: "Autauga" }), id: "01001" },
        { ...feature({ fips: "01003", NAME: "Baldwin" }), id: "ignored" },
      ),
      "fips",
      "NAME",
    );
    expect(fc.features.map((f) => f.properties)).toEqual([
      { id: "01001", name: "Autauga" },
      { id: "01003", name: "Baldwin" },
    ]);
  });

  it("drops features without polygons or ids", () => {
    const fc = buildCustomFeatures(
      collection(
        feature({ code: "501" }),
        feature({ code: "502" }, { type: "Point", coordinates: [0, 0] }),
        feature({ code: "" }),
        feature({ other: "504" }),
        feature(null),
        feature(
          { code: "505" },
          { type: "MultiPolygon", coordinates: [square.coordinates] },
        ),
      ),
      "code",
      undefined,
    );
    expect(fc.features.map((f) => f.properties?.id)).toEqual(["501", "505"]);
  });

  it("keeps the first of several features with the same id", () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    const fc = buildCustomFeatures(
      collection(
        feature({ code: "501", name: "first" }),
        feature({ code: "501", name: "second" }),
      ),
      "code",
      "name",
    );
    expect(fc.features.map((f) => f.properties?.name)).toEqual(["first"]);
    expect(warn).toHaveBeenCalledOnce();
    warn.mockRestore();
  });

  it("fails when no feature can be drawn", () => {
    expect(() =>
      buildCustomFeatures(collection(feature({ other: "1" })), "code", "name"),
    ).toThrow(/no polygons with a "code" property or a feature id/);
  });

  it("matches dimension values on the id", () => {
    const fc = buildCustomFeatures(
      collection(feature({ code: "0501" })),
      "code",
      undefined,
    );
    const custom = { file: "public/dma.geojson", id: "code" };
    const index = buildRegionIndex(
      fc.features,
      regionIndexKeys(custom, undefined),
    );
    expect(index.get("501")).toEqual(["0501"]);
  });
});

describe("isValidRegionBoundaries", () => {
  it("accepts built-in sets and complete custom boundaries", () => {
    expect(isValidRegionBoundaries("countries")).toBe(true);
    expect(isValidRegionBoundaries("dmas")).toBe(false);
    expect(isValidRegionBoundaries({ file: "public/a.json", id: "x" })).toBe(
      true,
    );
    expect(isValidRegionBoundaries({ file: "public/a.json" })).toBe(false);
    expect(isValidRegionBoundaries({ file: "", id: "x" })).toBe(false);
  });
});
