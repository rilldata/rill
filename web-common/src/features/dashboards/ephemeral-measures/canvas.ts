import type { EphemeralMeasureDef } from "./types";

// YAML shape of an ephemeral measure in a canvas component spec (snake_case).
export interface EphemeralMeasureSpec {
  name: string;
  display_name: string;
  expression: string;
  format_preset?: string;
  description?: string;
}

export function ephemeralSpecsToDefs(
  specs: EphemeralMeasureSpec[] | undefined,
): EphemeralMeasureDef[] | undefined {
  if (!specs?.length) return undefined;
  return specs
    .filter((spec) => spec?.name && spec?.expression)
    .map((spec) => ({
      name: spec.name,
      displayName: spec.display_name || spec.name,
      expression: spec.expression,
      ...(spec.format_preset ? { formatPreset: spec.format_preset } : {}),
      ...(spec.description ? { description: spec.description } : {}),
    }));
}

export function ephemeralDefsToSpecs(
  defs: EphemeralMeasureDef[] | undefined,
): EphemeralMeasureSpec[] {
  return (defs ?? []).map((def) => ({
    name: def.name,
    display_name: def.displayName,
    expression: def.expression,
    ...(def.formatPreset ? { format_preset: def.formatPreset } : {}),
    ...(def.description ? { description: def.description } : {}),
  }));
}

/**
 * Returns the component spec properties that reference the given measure,
 * with the reference removed, keyed the way `updateProperty` expects (an
 * `undefined` value deletes the property). Handles plain lists (`measures`,
 * `columns`), the single-measure `measure` property, chart field configs
 * (`y.field`, `y.fields`, `color.field`, ...) and per-measure entries such as
 * the pivot's `conditional_format`. Other properties are never touched even
 * if a value equals the name: `metrics_view`, `title`, dimension lists, or a
 * `comparison` list containing "delta". A field config whose only field was
 * the measure is dropped entirely, matching the inspector's own remove action.
 */
export function removeMeasureFromComponentSpec(
  spec: Record<string, unknown>,
  name: string,
): Record<string, unknown> {
  const changes: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(spec)) {
    if (typeof value === "string") {
      if (key === "measure" && value === name) changes[key] = undefined;
    } else if (Array.isArray(value)) {
      const kept = value.filter((item: unknown) =>
        MEASURE_LIST_KEYS.has(key)
          ? item !== name
          : !(isRecord(item) && item["measure"] === name),
      );
      if (kept.length !== value.length) changes[key] = kept;
    } else if (isRecord(value)) {
      const rawFields = Array.isArray(value["fields"])
        ? (value["fields"] as unknown[])
        : undefined;
      const fields = rawFields?.filter((f) => f !== name) ?? [];
      const fieldsChanged =
        rawFields !== undefined && fields.length !== rawFields.length;
      if (value["field"] !== name && !fieldsChanged) continue;
      const next: Record<string, unknown> = { ...value };
      if (fieldsChanged) {
        if (fields.length) next["fields"] = fields;
        else delete next["fields"];
      }
      if (next["field"] === name) {
        if (fields.length) next["field"] = fields[0];
        else delete next["field"];
      }
      changes[key] = "field" in next || "fields" in next ? next : undefined;
    }
  }
  return changes;
}

// Component spec lists whose string entries are measure names.
const MEASURE_LIST_KEYS = new Set(["measures", "columns"]);

function isRecord(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}
