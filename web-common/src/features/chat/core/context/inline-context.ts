export const INLINE_CHAT_CONTEXT_TAG = "chat-reference";

export enum InlineContextType {
  MetricsView = "metricsView",
  TimeRange = "timeRange",
  Measure = "measure",
  Dimension = "dimension",
  DimensionValues = "dimensionValues",
  Model = "model",
  Column = "column",
}

export type InlineContext = {
  type: InlineContextType;
  label?: string;
  value: string; // Main value for this context, used as an ID of sorts
  metricsView?: string;
  measure?: string;
  dimension?: string;
  timeRange?: string;
  values?: string[];
  model?: string;
  column?: string;
  columnType?: string; // TODO: is this needed here?
};

export function inlineContextsAreEqual(
  ctx1: InlineContext,
  ctx2: InlineContext,
) {
  const nonValuesAreEqual =
    ctx1.type === ctx2.type &&
    ctx1.metricsView === ctx2.metricsView &&
    ctx1.measure === ctx2.measure &&
    ctx1.dimension === ctx2.dimension &&
    ctx1.timeRange === ctx2.timeRange &&
    ctx1.model === ctx2.model &&
    ctx1.column === ctx2.column &&
    ctx1.columnType === ctx2.columnType;
  if (!nonValuesAreEqual) return false;
  if (!ctx1.values && !ctx2.values) return true;
  else if (!ctx1.values || !ctx2.values) return false;

  return (
    ctx1.values.length === ctx2.values.length &&
    ctx1.values.every((value, index) => value === ctx2.values![index])
  );
}

export function inlineContextIsWithin(src: InlineContext, tar: InlineContext) {
  if (src.type === tar.type) return false; // Equal types cannot be within each other, just equal
  switch (src.type) {
    case InlineContextType.MetricsView:
      return src.metricsView === tar.metricsView;
    case InlineContextType.Model:
      return src.model === tar.model;
  }
  return false;
}

export function normalizeInlineContext(ctx: InlineContext) {
  const normalisedContext = Object.fromEntries(
    Object.entries(ctx).filter(([, v]) => v !== null && v !== undefined),
  ) as InlineContext;

  // Fill in the `value` field based on the type of the context.`
  switch (normalisedContext.type) {
    case InlineContextType.MetricsView:
      normalisedContext.value = normalisedContext.metricsView!;
      break;

    case InlineContextType.Measure:
      normalisedContext.value = normalisedContext.measure!;
      break;

    case InlineContextType.Dimension:
      normalisedContext.value = normalisedContext.dimension!;
      break;

    case InlineContextType.TimeRange:
      normalisedContext.value = normalisedContext.timeRange!;
      break;

    case InlineContextType.Model:
      normalisedContext.value = normalisedContext.model!;
      break;

    case InlineContextType.Column:
      normalisedContext.value = normalisedContext.column!;
      break;
  }

  return normalisedContext;
}

// =============================================================================
// Utils for converting between different formats
// =============================================================================

export function convertContextToInlinePrompt(ctx: InlineContext) {
  const parts: string[] = [];

  for (const key in ctx) {
    const isComputedKey = key === "value" || key === "label";
    const isNonStringKey = key === "values";
    if (isComputedKey || isNonStringKey) continue;
    if (ctx[key] !== undefined) parts.push(`${key}="${ctx[key]}"`);
  }

  // TODO: dimension value support

  return `<${INLINE_CHAT_CONTEXT_TAG}>${parts.join(" ")}</${INLINE_CHAT_CONTEXT_TAG}>`;
}

const PARTS_REGEX = /(\w+?)="([^"]+?)"/g;

export function convertPromptValueToContext(
  contextValue: string,
): InlineContext | null {
  const parts = contextValue.matchAll(PARTS_REGEX);

  const ctx = <InlineContext>{};

  for (const [, key, value] of parts) {
    ctx[key] = value;
  }

  if (!ctx.type) return null;

  return normalizeInlineContext(ctx);
}

export function parseInlineAttr(content: string, key: string) {
  const match = new RegExp(`${key}="([^"]+?)"`).exec(content);
  if (!match) return null;
  return match[1];
}
