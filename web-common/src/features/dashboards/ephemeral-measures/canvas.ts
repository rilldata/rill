import type { EphemeralMeasureDef } from "./types";

// YAML shape of an ephemeral measure in a canvas component spec (snake_case).
export interface EphemeralMeasureSpec {
  name: string;
  display_name: string;
  expression: string;
  format_preset?: string;
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
  }));
}
