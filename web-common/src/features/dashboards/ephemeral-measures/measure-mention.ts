/**
 * Helpers for the "@" measure picker in the ephemeral measure expression
 * editor (MeasureExpressionInput.svelte).
 *
 * Typing "@" in the expression opens a picker of the referenceable measures,
 * searchable by display name as well as by name, so users do not need to know
 * a measure's field name to reference it. A picked measure is shown as a chip
 * with its display name, like a mention in the AI chat, and serializes to its
 * (quoted if needed) name in the expression.
 */
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { formatMeasureRef } from "./expression-parser";

// A "@" only starts a mention when it does not directly follow an identifier or
// a quoted name, so "@" pasted in the middle of a name is not picked up.
export const IDENTIFIER_CHAR = /[A-Za-z0-9_"]/;

export type ExpressionToken =
  | { type: "text"; text: string }
  | { type: "measure"; name: string; displayName: string };

// A quoted identifier, or a whole bare identifier that is not a function call.
const REFERENCE_REGEX = /"(?:[^"]|"")+"|\b[A-Za-z_][A-Za-z0-9_]*\b(?!\s*\()/g;

/**
 * Splits an expression into plain text and references to the given measures,
 * so an existing expression can be shown with measure chips. Names that are
 * not a known measure (functions, literals, typos) stay text.
 */
export function tokenizeMeasureExpression(
  expression: string,
  measures: MetricsViewSpecMeasure[],
): ExpressionToken[] {
  const byName = new Map(measures.map((mes) => [mes.name ?? "", mes]));
  const tokens: ExpressionToken[] = [];
  let last = 0;
  const pushText = (text: string) => {
    if (!text) return;
    const prev = tokens[tokens.length - 1];
    if (prev?.type === "text") {
      prev.text += text;
    } else {
      tokens.push({ type: "text", text });
    }
  };
  for (const match of expression.matchAll(REFERENCE_REGEX)) {
    const raw = match[0];
    const name = raw.startsWith('"')
      ? raw.slice(1, -1).replace(/""/g, '"')
      : raw;
    const measure = byName.get(name);
    pushText(expression.slice(last, match.index));
    if (measure) {
      tokens.push({
        type: "measure",
        name,
        displayName: measure.displayName || name,
      });
    } else {
      pushText(raw);
    }
    last = match.index + raw.length;
  }
  pushText(expression.slice(last));
  return tokens;
}

/**
 * Returns the measures matching the mention query by display name or name,
 * keeping the metrics view's order. An empty query matches everything.
 */
export function filterMeasureMentions(
  measures: MetricsViewSpecMeasure[],
  query: string,
): MetricsViewSpecMeasure[] {
  const needle = query.trim().toLowerCase();
  if (!needle) return measures;
  return measures.filter(
    (mes) =>
      (mes.displayName ?? "").toLowerCase().includes(needle) ||
      (mes.name ?? "").toLowerCase().includes(needle),
  );
}

/**
 * Serializes the editor's tokens back to an expression.
 */
export function serializeExpressionTokens(tokens: ExpressionToken[]): string {
  return tokens
    .map((token) =>
      token.type === "text" ? token.text : formatMeasureRef(token.name),
    )
    .join("");
}
