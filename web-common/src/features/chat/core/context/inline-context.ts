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
  // Main value for this context, used to run search and forms a unique identifier with type
  value: string;
  metricsView?: string;
  measure?: string;
  dimension?: string;
  timeRange?: string;
  values?: string[];
  model?: string;
  column?: string;
  columnType?: string; // TODO: is this needed here?
};

export function getIdForContext(ctx: InlineContext) {
  const parentIdPart = ctx.metricsView ?? ctx.model;
  return `${ctx.type}__${ctx.value}${parentIdPart ? `__${parentIdPart}` : ""}`;
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

    case InlineContextType.DimensionValues:
      normalisedContext.value = normalisedContext.values!.join(",");
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
