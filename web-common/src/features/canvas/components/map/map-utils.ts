import mapboxgl from "mapbox-gl";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import {
  mouseLocationToBoundingRect,
  placeElement,
} from "@rilldata/web-common/lib/place-element";

// ── GeoJSON transformation ──────────────────────────────────────

interface TransformOptions {
  geoDimension: string;
  colorMeasure: string | null;
  sizeMeasure: string | undefined;
  tooltipDimension: string | undefined;
}

export function transformToGeoJSON(
  data: V1MetricsViewAggregationResponseDataItem[],
  opts: TransformOptions,
): GeoJSON.FeatureCollection {
  const features: GeoJSON.Feature[] = [];

  for (const row of data) {
    const geoValue = row[opts.geoDimension];
    if (!geoValue) continue;

    let geometry: GeoJSON.Geometry | null = null;

    if (typeof geoValue === "string") {
      try {
        const parsed = JSON.parse(geoValue);
        if (parsed?.type && parsed?.coordinates) {
          geometry = parsed as GeoJSON.Geometry;
        } else if (parsed?.type === "Feature" && parsed?.geometry) {
          geometry = parsed.geometry as GeoJSON.Geometry;
        }
      } catch {
        continue;
      }
    } else if (Array.isArray(geoValue)) {
      // DuckDB spatial types are serialized as [x, y] = [lon, lat],
      // which is already the order GeoJSON/Mapbox expect — pass through.
      if (
        geoValue.length === 2 &&
        typeof geoValue[0] === "number" &&
        typeof geoValue[1] === "number"
      ) {
        const [lon, lat] = geoValue as [number, number];
        geometry = { type: "Point", coordinates: [lon, lat] };
      } else if (Array.isArray(geoValue[0]) && Array.isArray(geoValue[0][0])) {
        geometry = {
          type: "Polygon",
          coordinates: geoValue as number[][][],
        };
      }
    }

    if (!geometry) continue;

    const properties: Record<string, unknown> = {};
    if (opts.colorMeasure && row[opts.colorMeasure] != null) {
      properties[opts.colorMeasure] = Number(row[opts.colorMeasure]);
    }
    if (opts.sizeMeasure && row[opts.sizeMeasure] != null) {
      properties[opts.sizeMeasure] = Number(row[opts.sizeMeasure]);
    }
    if (opts.tooltipDimension && row[opts.tooltipDimension] != null) {
      properties[opts.tooltipDimension] = row[opts.tooltipDimension];
    }

    features.push({ type: "Feature", geometry, properties });
  }

  return { type: "FeatureCollection", features };
}

// ── Bounds calculation ──────────────────────────────────────────

function extendBoundsWithCoord(bounds: mapboxgl.LngLatBounds, coord: number[]) {
  const [lng, lat] = coord;
  if (
    typeof lng === "number" &&
    typeof lat === "number" &&
    lng >= -180 &&
    lng <= 180 &&
    lat >= -90 &&
    lat <= 90
  ) {
    bounds.extend([lng, lat]);
  }
}

export function calculateBounds(
  features: GeoJSON.Feature[],
): mapboxgl.LngLatBounds | null {
  if (features.length === 0) return null;

  const bounds = new mapboxgl.LngLatBounds();
  let hasValidCoord = false;

  for (const feature of features) {
    const geom = feature.geometry;
    switch (geom.type) {
      case "Point":
        extendBoundsWithCoord(bounds, geom.coordinates);
        hasValidCoord = true;
        break;
      case "MultiPoint":
      case "LineString":
        for (const coord of geom.coordinates) {
          extendBoundsWithCoord(bounds, coord);
        }
        hasValidCoord = true;
        break;
      case "Polygon":
      case "MultiLineString":
        for (const ring of geom.coordinates) {
          for (const coord of ring) {
            extendBoundsWithCoord(bounds, coord);
          }
        }
        hasValidCoord = true;
        break;
      case "MultiPolygon":
        for (const polygon of geom.coordinates) {
          for (const ring of polygon) {
            for (const coord of ring) {
              extendBoundsWithCoord(bounds, coord);
            }
          }
        }
        hasValidCoord = true;
        break;
    }
  }

  return hasValidCoord ? bounds : null;
}

/**
 * Calculates the bounds of features that may cross the antimeridian, such as Russia or Fiji.
 * Longitudes are also measured with the western hemisphere shifted east by 360°,
 * and whichever measurement spans fewer degrees is used,
 * so the camera fits the shape instead of zooming out to the whole world.
 * The returned east edge can exceed 180°, which Mapbox handles.
 */
export function calculateWrappedBounds(
  features: GeoJSON.Feature[],
): mapboxgl.LngLatBoundsLike | null {
  let south = Infinity;
  let north = -Infinity;
  let west = Infinity;
  let east = -Infinity;
  let shiftedWest = Infinity;
  let shiftedEast = -Infinity;

  for (const feature of features) {
    forEachCoordinate(feature.geometry, ([lng, lat]) => {
      south = Math.min(south, lat);
      north = Math.max(north, lat);
      west = Math.min(west, lng);
      east = Math.max(east, lng);
      const shifted = lng < 0 ? lng + 360 : lng;
      shiftedWest = Math.min(shiftedWest, shifted);
      shiftedEast = Math.max(shiftedEast, shifted);
    });
  }
  if (south === Infinity) return null;

  if (shiftedEast - shiftedWest < east - west) {
    return [
      [shiftedWest, south],
      [shiftedEast, north],
    ];
  }
  return [
    [west, south],
    [east, north],
  ];
}

function forEachCoordinate(
  geometry: GeoJSON.Geometry,
  fn: (coord: number[]) => void,
) {
  switch (geometry.type) {
    case "Point":
      fn(geometry.coordinates);
      break;
    case "MultiPoint":
    case "LineString":
      geometry.coordinates.forEach(fn);
      break;
    case "Polygon":
    case "MultiLineString":
      geometry.coordinates.forEach((ring) => ring.forEach(fn));
      break;
    case "MultiPolygon":
      geometry.coordinates.forEach((polygon) =>
        polygon.forEach((ring) => ring.forEach(fn)),
      );
      break;
    case "GeometryCollection":
      geometry.geometries.forEach((g) => forEachCoordinate(g, fn));
      break;
  }
}

// ── Polygon detection ───────────────────────────────────────────

export function detectPolygonMode(
  rows: V1MetricsViewAggregationResponseDataItem[],
  geoDimension: string,
): boolean {
  if (!rows.length || !geoDimension) return false;
  return rows.some((row) => {
    const v = row[geoDimension];
    if (typeof v === "string") {
      try {
        const parsed = JSON.parse(v);
        const t = parsed?.type ?? parsed?.geometry?.type;
        return t === "Polygon" || t === "MultiPolygon";
      } catch {
        return false;
      }
    }
    return Array.isArray(v) && Array.isArray(v[0]) && Array.isArray(v[0][0]);
  });
}

// ── Tooltip ─────────────────────────────────────────────────────

const MAP_TOOLTIP_ID = "rill-map-tooltip";

function escapeHTML(value: unknown): string {
  return String(value).replace(/&/g, "&amp;").replace(/</g, "&lt;");
}

export interface TooltipRow {
  label: string;
  value: string;
}

export function buildTooltipHTML(
  title: string | undefined,
  rows: TooltipRow[],
): string | null {
  let html = "";
  if (title) {
    html += `<h2>${escapeHTML(title)}</h2>`;
  }
  if (rows.length > 0) {
    const cells = rows
      .map(
        (r) =>
          `<tr><td class="key">${escapeHTML(r.label)}</td><td class="value">${escapeHTML(r.value)}</td></tr>`,
      )
      .join("");
    html += `<table><tbody>${cells}</tbody></table>`;
  }
  return html || null;
}

export function showTooltip(
  event: mapboxgl.MapMouseEvent,
  html: string | null,
) {
  removeTooltip();
  if (!html) return;

  const el = document.createElement("div");
  el.setAttribute("id", MAP_TOOLTIP_ID);
  el.innerHTML = html;
  document.body.appendChild(el);

  const parentRect = mouseLocationToBoundingRect({
    x: event.originalEvent.clientX,
    y: event.originalEvent.clientY,
  });
  const elementRect = el.getBoundingClientRect();

  const [leftPos, topPos] = placeElement({
    location: "right",
    alignment: "middle",
    distance: 12,
    pad: 8,
    parentPosition: parentRect,
    elementPosition: elementRect,
  });

  el.setAttribute("style", `top: ${topPos}px; left: ${leftPos}px`);
}

export function removeTooltip() {
  const el = document.getElementById(MAP_TOOLTIP_ID);
  if (el) el.remove();
}
