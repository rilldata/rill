import mapboxgl from "mapbox-gl";
import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import {
  mouseLocationToBoundingRect,
  placeElement,
} from "@rilldata/web-common/lib/place-element";
import { justEnoughPrecision } from "@rilldata/web-common/lib/formatters";

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
      // DuckDB spatial types return [lat, lon] — swap to [lon, lat]
      if (
        geoValue.length === 2 &&
        typeof geoValue[0] === "number" &&
        typeof geoValue[1] === "number"
      ) {
        const [lat, lon] = geoValue as [number, number];
        geometry = { type: "Point", coordinates: [lon, lat] };
      } else if (Array.isArray(geoValue[0]) && Array.isArray(geoValue[0][0])) {
        const coordinates = (geoValue as number[][][]).map((ring) =>
          ring.map(([lat, lon]) => [lon, lat]),
        );
        geometry = { type: "Polygon", coordinates };
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

interface TooltipContext {
  tooltipDimension: string | undefined;
  colorMeasure: string | null;
  sizeMeasure: string | undefined;
  getDisplayName: (name: string) => string;
}

export function buildTooltipHTML(
  properties: Record<string, unknown> | null,
  ctx: TooltipContext,
): string | null {
  if (!properties) return null;

  const { tooltipDimension, colorMeasure: cm, sizeMeasure: sm } = ctx;
  if (!tooltipDimension && !cm && !sm) return null;

  let html = "";

  if (tooltipDimension && properties[tooltipDimension] != null) {
    html += `<h2>${escapeHTML(properties[tooltipDimension])}</h2>`;
  }

  const rows: string[] = [];
  if (cm && properties[cm] != null) {
    const val =
      typeof properties[cm] === "number"
        ? justEnoughPrecision(properties[cm] as number)
        : String(properties[cm]);
    rows.push(
      `<tr><td class="key">${escapeHTML(ctx.getDisplayName(cm))}</td><td class="value">${escapeHTML(val)}</td></tr>`,
    );
  }
  if (sm && sm !== cm && properties[sm] != null) {
    const val =
      typeof properties[sm] === "number"
        ? justEnoughPrecision(properties[sm] as number)
        : String(properties[sm]);
    rows.push(
      `<tr><td class="key">${escapeHTML(ctx.getDisplayName(sm))}</td><td class="value">${escapeHTML(val)}</td></tr>`,
    );
  }

  if (rows.length > 0) {
    html += `<table><tbody>${rows.join("")}</tbody></table>`;
  }

  return html || null;
}

export function showTooltip(
  event: mapboxgl.MapMouseEvent & { features?: GeoJSON.Feature[] },
  ctx: TooltipContext,
) {
  removeTooltip();

  const feature = event.features?.[0];
  if (!feature) return;

  const html = buildTooltipHTML(
    feature.properties as Record<string, unknown> | null,
    ctx,
  );
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
