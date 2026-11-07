import type {
  PivotFormatRule,
  PivotFormatRuleOperator,
  PivotMeasureFormatting,
} from "./types";

/**
 * Compact serialization of per-measure conditional formatting for the stateful
 * URL (and the explore preset, which stores the same string).
 *
 * Grammar:
 *   param  := entry (";" entry)*
 *   entry  := measure ":" ("heatmap" | "data_bar") ":" scheme
 *           | measure ":" "rules" ":" rule ("|" rule)*
 *   rule   := op "," value ["," value2] "," color
 *
 * Example: `margin:rules:lt,0,negative|gte,0.4,positive;revenue:heatmap:greens`
 * Hex rule colors are stored without the leading "#" to keep URLs clean.
 */

const OPERATORS = new Set<PivotFormatRuleOperator>([
  "gt",
  "gte",
  "lt",
  "lte",
  "eq",
  "between",
]);

export function toPivotFormattingParam(
  measureFormatting: Record<string, PivotMeasureFormatting> | undefined,
): string {
  if (!measureFormatting) return "";
  return Object.entries(measureFormatting)
    .map(([measure, fmt]) => {
      if (fmt.mode === "rules") {
        const rules = fmt.rules.map(ruleToString).join("|");
        return `${measure}:rules:${rules}`;
      }
      return [measure, fmt.mode, fmt.scheme].join(":");
    })
    .join(";");
}

function ruleToString(rule: PivotFormatRule): string {
  const parts: string[] = [rule.operator, String(rule.value)];
  if (rule.operator === "between") {
    parts.push(String(rule.value2 ?? rule.value));
  }
  parts.push(rule.color.replace(/^#/, ""));
  return parts.join(",");
}

export function fromPivotFormattingParam(param: string): {
  measureFormatting: Record<string, PivotMeasureFormatting> | undefined;
  invalidEntries: string[];
} {
  const measureFormatting: Record<string, PivotMeasureFormatting> = {};
  const invalidEntries: string[] = [];

  for (const entry of param.split(";")) {
    if (!entry) continue;
    const parsed = parseEntry(entry);
    if (parsed) {
      measureFormatting[parsed.measure] = parsed.fmt;
    } else {
      invalidEntries.push(entry);
    }
  }

  return {
    measureFormatting: Object.keys(measureFormatting).length
      ? measureFormatting
      : undefined,
    invalidEntries,
  };
}

function parseEntry(
  entry: string,
): { measure: string; fmt: PivotMeasureFormatting } | undefined {
  const [measure, mode, ...rest] = entry.split(":");
  if (!measure || rest.length === 0) return undefined;

  if (mode === "heatmap" || mode === "data_bar") {
    const [scheme] = rest;
    if (!scheme || rest.length > 1) return undefined;
    return { measure, fmt: { mode, scheme } };
  }

  if (mode === "rules") {
    const rules: PivotFormatRule[] = [];
    for (const ruleStr of rest.join(":").split("|")) {
      const rule = parseRule(ruleStr);
      if (!rule) return undefined;
      rules.push(rule);
    }
    if (rules.length === 0) return undefined;
    return { measure, fmt: { mode: "rules", rules } };
  }

  return undefined;
}

function parseRule(ruleStr: string): PivotFormatRule | undefined {
  const parts = ruleStr.split(",");
  const operator = parts[0] as PivotFormatRuleOperator;
  if (!OPERATORS.has(operator)) return undefined;

  const expectedParts = operator === "between" ? 4 : 3;
  if (parts.length !== expectedParts) return undefined;

  const value = Number(parts[1]);
  if (!Number.isFinite(value)) return undefined;

  const color = parts[expectedParts - 1];
  if (!color) return undefined;

  if (operator === "between") {
    const value2 = Number(parts[2]);
    if (!Number.isFinite(value2)) return undefined;
    return { operator, value, value2, color };
  }

  return { operator, value, color };
}
