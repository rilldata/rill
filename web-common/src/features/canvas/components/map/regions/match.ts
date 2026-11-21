import type { V1MetricsViewAggregationResponseDataItem } from "@rilldata/web-common/runtime-client";
import type * as GeoJSON from "geojson";

/** Maps normalized dimension values to the ids of the regions they refer to. */
export type RegionIndex = Map<string, string[]>;

export interface RegionValue {
  value: number;
  /** The dimension value as returned by the query, e.g. "USA". */
  label: string;
}

export interface RegionMatch {
  /** Measure value per region id. */
  values: Map<string, RegionValue>;
  /** Distinct dimension values that did not match any region. */
  unmatched: string[];
}

/**
 * Normalizes a region name or code for matching:
 * case, accents, punctuation and whitespace are ignored,
 * and leading zeros are dropped from numeric codes so that "036" matches "36".
 */
export function normalizeRegionValue(value: unknown): string {
  const normalized = String(value)
    .normalize("NFD")
    .replace(/[̀-ͯ]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]/g, "");
  return /^\d+$/.test(normalized)
    ? normalized.replace(/^0+(?=\d)/, "")
    : normalized;
}

/**
 * Indexes regions by the given keys.
 * The `name` key also indexes each region's aliases.
 * When two keys produce the same value, the key listed first wins,
 * while regions that share a value for the same key (such as territories coded as their parent country) all match it.
 */
export function buildRegionIndex(
  features: GeoJSON.Feature[],
  keys: string[],
): RegionIndex {
  const index: RegionIndex = new Map();
  for (const key of keys) {
    const indexedByKey = new Set<string>();
    for (const feature of features) {
      const id = feature.properties?.id as string | undefined;
      if (!id) continue;
      for (const value of featureValues(feature, key)) {
        const normalized = normalizeRegionValue(value);
        if (!normalized) continue;
        const ids = index.get(normalized);
        if (!ids) {
          index.set(normalized, [id]);
          indexedByKey.add(normalized);
        } else if (indexedByKey.has(normalized) && !ids.includes(id)) {
          ids.push(id);
        }
      }
    }
  }
  return index;
}

/**
 * Joins aggregation rows to regions through the dimension value.
 * When several rows match the same region (e.g. both "US" and "USA" are present), the first row wins.
 */
export function matchRegions(
  rows: V1MetricsViewAggregationResponseDataItem[],
  dimension: string,
  measure: string,
  index: RegionIndex,
): RegionMatch {
  const values = new Map<string, RegionValue>();
  const unmatched = new Set<string>();
  for (const row of rows) {
    const label = row[dimension];
    const value = row[measure];
    if (label == null || value == null) continue;

    const ids = index.get(normalizeRegionValue(label));
    if (!ids) {
      unmatched.add(String(label));
      continue;
    }
    for (const id of ids) {
      if (!values.has(id)) {
        values.set(id, { value: Number(value), label: String(label) });
      }
    }
  }
  return { values, unmatched: [...unmatched] };
}

function featureValues(feature: GeoJSON.Feature, key: string): unknown[] {
  const props = feature.properties ?? {};
  if (key !== "name") return props[key] != null ? [props[key]] : [];

  const aliases =
    typeof props.aliases === "string" && props.aliases
      ? props.aliases.split("|")
      : [];
  return [props.name, ...aliases];
}
