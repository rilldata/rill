import type { EphemeralMeasureDef } from "./types";

/**
 * The ad-hoc measure library persists every definition a user creates for a
 * metrics view in localStorage, keyed by the metrics view rather than the
 * explore. Any explore on the same metrics view restores the library on load,
 * so a definition survives without a bookmark and is not lost when it is
 * hidden: the URL only carries definitions the explore state references (see
 * `referencedEphemeralMeasures`), which keeps shared links short.
 *
 * Deletions must be explicit (`syncEphemeralMeasureLibrary` with the previous
 * definitions); loading never removes entries, since an explore may expose a
 * subset of the metrics view's measures and drop definitions it cannot use.
 */

function getKeyForLibrary(
  metricsViewName: string,
  storageNamespacePrefix: string | undefined,
) {
  return `rill:app:adhoc-measures:${storageNamespacePrefix ?? ""}${metricsViewName}`.toLowerCase();
}

export function loadEphemeralMeasureLibrary(
  metricsViewName: string,
  storageNamespacePrefix: string | undefined,
): EphemeralMeasureDef[] {
  try {
    const raw = localStorage.getItem(
      getKeyForLibrary(metricsViewName, storageNamespacePrefix),
    );
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter(isEphemeralMeasureDef);
  } catch {
    return [];
  }
}

export function saveEphemeralMeasureLibrary(
  metricsViewName: string,
  storageNamespacePrefix: string | undefined,
  defs: EphemeralMeasureDef[],
) {
  try {
    const key = getKeyForLibrary(metricsViewName, storageNamespacePrefix);
    if (!defs.length) {
      localStorage.removeItem(key);
      return;
    }
    localStorage.setItem(key, JSON.stringify(defs));
  } catch {
    // no-op: storage may be unavailable (private mode, embeds, quota)
  }
}

/**
 * Upserts the given definitions into the library without removing anything.
 * Used on load, where a missing definition does not mean it was deleted.
 */
export function upsertIntoEphemeralMeasureLibrary(
  metricsViewName: string,
  storageNamespacePrefix: string | undefined,
  defs: EphemeralMeasureDef[] | undefined,
) {
  if (!defs?.length) return;
  const library = loadEphemeralMeasureLibrary(
    metricsViewName,
    storageNamespacePrefix,
  );
  saveEphemeralMeasureLibrary(
    metricsViewName,
    storageNamespacePrefix,
    mergeEphemeralMeasureDefs(library, defs, defs),
  );
}

/**
 * Applies a state transition to the library: definitions in `next` are
 * upserted, and definitions present in `previous` but absent from `next`
 * were deleted by the user and are removed.
 */
export function syncEphemeralMeasureLibrary(
  metricsViewName: string,
  storageNamespacePrefix: string | undefined,
  previous: EphemeralMeasureDef[] | undefined,
  next: EphemeralMeasureDef[] | undefined,
) {
  const nextNames = new Set(next?.map((def) => def.name) ?? []);
  const removed = (previous ?? []).filter((def) => !nextNames.has(def.name));
  if (!removed.length && !next?.length) return;

  const library = loadEphemeralMeasureLibrary(
    metricsViewName,
    storageNamespacePrefix,
  );
  const removedNames = new Set(removed.map((def) => def.name));
  saveEphemeralMeasureLibrary(
    metricsViewName,
    storageNamespacePrefix,
    mergeEphemeralMeasureDefs(
      library.filter((def) => !removedNames.has(def.name)),
      next ?? [],
      next ?? [],
    ),
  );
}

/**
 * Unions two definition lists by name, keeping `base` order and appending
 * unseen entries from `extra`. Entries from `override` replace same-named
 * entries from `base` (used so an edit wins over the stored copy).
 */
export function mergeEphemeralMeasureDefs(
  base: EphemeralMeasureDef[],
  extra: EphemeralMeasureDef[],
  override: EphemeralMeasureDef[] = [],
): EphemeralMeasureDef[] {
  const overrides = new Map(override.map((def) => [def.name, def]));
  const seen = new Set<string>();
  const merged: EphemeralMeasureDef[] = [];
  for (const def of [...base, ...extra]) {
    if (seen.has(def.name)) continue;
    seen.add(def.name);
    merged.push(overrides.get(def.name) ?? def);
  }
  return merged;
}

function isEphemeralMeasureDef(value: unknown): value is EphemeralMeasureDef {
  if (!value || typeof value !== "object") return false;
  const def = value as Record<string, unknown>;
  return (
    typeof def.name === "string" &&
    !!def.name &&
    typeof def.displayName === "string" &&
    !!def.displayName &&
    typeof def.expression === "string" &&
    !!def.expression &&
    (def.formatPreset === undefined || typeof def.formatPreset === "string")
  );
}
