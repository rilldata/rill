import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery } from "@tanstack/svelte-query";
import type * as GeoJSON from "geojson";
import { derived } from "svelte/store";
import { feature } from "topojson-client";
import type { Topology } from "topojson-specification";
import type { MapComponent, MapSpec } from "..";
import {
  AUTO_REGION_KEY,
  BOUNDARY_SETS,
  isRegionBoundaries,
  type RegionBoundaries,
} from "./registry";

/** Boundaries from a GeoJSON or TopoJSON file under one of the project's public paths. */
export interface CustomRegionBoundaries {
  /** Path of the file relative to the project root, such as `public/dma.topojson`. */
  file: string;
  /** Feature property matched against the dimension's values. */
  id: string;
  /** Feature property shown as the region name. Defaults to `id`. */
  name?: string;
}

/** Value of the `regions.boundaries` property: a built-in set or a custom file. */
export type RegionBoundariesConfig = RegionBoundaries | CustomRegionBoundaries;

export function isCustomRegionBoundaries(
  value: unknown,
): value is CustomRegionBoundaries {
  if (!value || typeof value !== "object") return false;
  const { file, id } = value as Partial<CustomRegionBoundaries>;
  return typeof file === "string" && !!file && typeof id === "string" && !!id;
}

/** Whether the boundaries are complete enough to load. */
export function isValidRegionBoundaries(
  value: unknown,
): value is RegionBoundariesConfig {
  return isRegionBoundaries(value) || isCustomRegionBoundaries(value);
}

/** Identifies the boundaries, e.g. to tell when the drawn regions change. */
export function regionBoundariesKey(boundaries: RegionBoundariesConfig) {
  return isRegionBoundaries(boundaries)
    ? boundaries
    : `${boundaries.file}::${boundaries.id}::${boundaries.name ?? ""}`;
}

/**
 * Feature properties the dimension values are matched against.
 * Custom boundaries are always matched on `id`, which holds the property named in `boundaries.id`.
 */
export function regionIndexKeys(
  boundaries: RegionBoundariesConfig,
  key: string | undefined,
): string[] {
  if (!isRegionBoundaries(boundaries)) return ["id"];
  if (key && key !== AUTO_REGION_KEY) return [key];
  return BOUNDARY_SETS[boundaries].keys.map((k) => k.name);
}

/**
 * Loads the regions layer's boundaries as features with a unique `properties.id` and a display `properties.name`.
 * Boundaries don't depend on filters or the time range, and custom files only change on deploy,
 * so they are loaded once and never refetched while cached;
 * after a page reload, the runtime's ETag makes fetching a custom file again a cheap revalidation.
 */
export function createRegionBoundariesQuery(
  runtimeClient: RuntimeClient,
  component: MapComponent,
) {
  const { specStore, dataEnabled } = component;

  const queryOptionsStore = derived(
    [specStore, dataEnabled],
    ([specVal, $visible]) => {
      const boundaries = (specVal as Partial<MapSpec> | undefined)?.regions
        ?.boundaries;
      const valid = isValidRegionBoundaries(boundaries);
      return {
        queryKey: [
          "map-region-boundaries",
          runtimeClient.instanceId,
          valid ? regionBoundariesKey(boundaries) : "",
        ],
        queryFn: () => {
          if (isRegionBoundaries(boundaries))
            return BOUNDARY_SETS[boundaries].load();
          return loadCustomBoundaries(
            runtimeClient,
            boundaries as CustomRegionBoundaries,
          );
        },
        enabled: $visible && valid,
        staleTime: Infinity,
        // Boundaries are large and replaced as a whole, so diffing them against the previous result is wasted work.
        structuralSharing: false,
        retry: false,
      };
    },
  );

  return createQuery(queryOptionsStore);
}

async function loadCustomBoundaries(
  runtimeClient: RuntimeClient,
  boundaries: CustomRegionBoundaries,
): Promise<GeoJSON.FeatureCollection> {
  const { file, id, name } = boundaries;
  const headers: Record<string, string> = {};
  const jwt = runtimeClient.getJwt();
  if (jwt) headers["Authorization"] = `Bearer ${jwt}`;

  const path = file.replace(/^\/+/, "");
  const response = await fetch(
    `${runtimeClient.host}/v1/instances/${runtimeClient.instanceId}/assets/${path}`,
    { headers },
  );
  if (!response.ok) {
    let message = response.statusText || `HTTP ${response.status}`;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // Keep the status text.
    }
    if (response.status === 403) {
      message += "; the file must be under one of the project's public paths";
    }
    throw new Error(`Failed to load boundaries from "${file}": ${message}`);
  }

  let json: unknown;
  try {
    json = await response.json();
  } catch {
    throw new Error(`Failed to load boundaries from "${file}": invalid JSON`);
  }
  try {
    return buildCustomFeatures(decodeBoundaries(json), id, name);
  } catch (err) {
    throw new Error(
      `Failed to load boundaries from "${file}": ${(err as Error).message}`,
    );
  }
}

/** Decodes a GeoJSON `FeatureCollection` or a TopoJSON topology with a single object. */
export function decodeBoundaries(json: unknown): GeoJSON.FeatureCollection {
  const type = (json as { type?: unknown } | null)?.type;
  if (type === "FeatureCollection") return json as GeoJSON.FeatureCollection;
  if (type !== "Topology") {
    throw new Error(
      "expected a GeoJSON FeatureCollection or a TopoJSON Topology",
    );
  }

  const topology = json as Topology;
  const names = Object.keys(topology.objects ?? {});
  if (names.length !== 1) {
    throw new Error(
      `expected a TopoJSON file with one object, found ${names.length}${names.length ? ` (${names.join(", ")})` : ""}`,
    );
  }
  const decoded = feature(topology, topology.objects[names[0]]);
  return decoded.type === "FeatureCollection"
    ? decoded
    : { type: "FeatureCollection", features: [decoded] };
}

/**
 * Builds features with a unique `properties.id` from the `id` property and a display `properties.name` from the `name` property.
 * When a feature lacks the `id` property, its top-level GeoJSON `id` is used instead,
 * since many files keep codes such as FIPS there.
 * Features without polygons or without an id are dropped,
 * and only the first feature with a given id is kept, since feature state is keyed by id.
 */
export function buildCustomFeatures(
  collection: GeoJSON.FeatureCollection,
  idProperty: string,
  nameProperty: string | undefined,
): GeoJSON.FeatureCollection {
  const features: GeoJSON.Feature[] = [];
  const seen = new Set<string>();
  const duplicates = new Set<string>();
  for (const f of collection.features ?? []) {
    const geometryType = f.geometry?.type;
    if (geometryType !== "Polygon" && geometryType !== "MultiPolygon") continue;

    const rawId: unknown = f.properties?.[idProperty] ?? f.id;
    if (rawId == null || rawId === "") continue;
    const id = String(rawId);
    if (seen.has(id)) {
      duplicates.add(id);
      continue;
    }
    seen.add(id);

    const rawName: unknown = nameProperty
      ? f.properties?.[nameProperty]
      : undefined;
    features.push({
      type: "Feature",
      geometry: f.geometry,
      properties: {
        id,
        name: rawName == null || rawName === "" ? id : String(rawName),
      },
    });
  }

  if (features.length === 0) {
    throw new Error(
      `no polygons with a "${idProperty}" property or a feature id`,
    );
  }
  if (duplicates.size > 0) {
    console.warn(
      `Map boundaries: kept the first of several regions with the same "${idProperty}": ${[...duplicates].slice(0, 10).join(", ")}`,
    );
  }
  return { type: "FeatureCollection", features };
}
