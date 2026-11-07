import type { EphemeralMeasureDef } from "./types";
import { MAX_EPHEMERAL_MEASURES } from "./validation";

/**
 * Compact serialization of ephemeral measure definitions for the stateful URL
 * (and the explore preset, which stores the same string).
 *
 * Grammar:
 *   param := entry (";" entry)*
 *   entry := name ":" enc(displayName) ":" enc(expression)
 *            [":" formatPreset [":" enc(description)]]
 *
 * `enc` is encodeURIComponent applied per field: it encodes ":" and ";" so
 * splitting is unambiguous. `name` is a validated slug and needs no encoding.
 * `formatPreset` is left empty when only a description follows.
 *
 * Example: `profit:Profit:revenue%20-%20cost;arpu:ARPU:revenue%20%2F%20users:currency_usd;m:Margin:profit%20%2F%20revenue::Profit%20per%20dollar`
 */

export function toEphemeralMeasuresParam(
  defs: EphemeralMeasureDef[] | undefined,
): string {
  if (!defs?.length) return "";
  // Mirrors the parse-side cap so a written param always reads back whole.
  return defs
    .slice(0, MAX_EPHEMERAL_MEASURES)
    .map((def) => {
      const parts = [
        def.name,
        encodeURIComponent(def.displayName),
        encodeURIComponent(def.expression),
      ];
      if (def.formatPreset || def.description) {
        parts.push(def.formatPreset ?? "");
      }
      if (def.description) parts.push(encodeURIComponent(def.description));
      return parts.join(":");
    })
    .join(";");
}

export function fromEphemeralMeasuresParam(param: string): {
  ephemeralMeasures: EphemeralMeasureDef[];
  invalidEntries: string[];
} {
  const ephemeralMeasures: EphemeralMeasureDef[] = [];
  const invalidEntries: string[] = [];
  const seen = new Set<string>();

  for (const entry of param.split(";")) {
    if (!entry) continue;
    const parts = entry.split(":");
    if (parts.length < 3 || parts.length > 5) {
      invalidEntries.push(entry);
      continue;
    }
    const [name, encDisplayName, encExpression, formatPreset, encDescription] =
      parts;
    let displayName: string;
    let expression: string;
    let description: string | undefined;
    try {
      displayName = decodeURIComponent(encDisplayName);
      expression = decodeURIComponent(encExpression);
      description = encDescription
        ? decodeURIComponent(encDescription)
        : undefined;
    } catch {
      invalidEntries.push(entry);
      continue;
    }
    if (!name || !displayName || !expression || seen.has(name)) {
      invalidEntries.push(entry);
      continue;
    }
    if (ephemeralMeasures.length >= MAX_EPHEMERAL_MEASURES) {
      invalidEntries.push(entry);
      continue;
    }
    seen.add(name);
    ephemeralMeasures.push({
      name,
      displayName,
      expression,
      ...(formatPreset ? { formatPreset } : {}),
      ...(description ? { description } : {}),
    });
  }

  return { ephemeralMeasures, invalidEntries };
}
