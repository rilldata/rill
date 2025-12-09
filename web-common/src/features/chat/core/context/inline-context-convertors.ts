import {
  INLINE_CHAT_CONTEXT_TAG,
  type InlineContext,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export function convertContextToInlinePrompt(ctx: InlineContext) {
  const parts: string[] = [`type="${ctx.type}"`];

  if (ctx.metricsView) {
    parts.push(`metricsView="${ctx.metricsView}"`);
  }
  if (ctx.measure) {
    parts.push(`measure="${ctx.measure}"`);
  }
  if (ctx.dimension) {
    parts.push(`dimension="${ctx.dimension}"`);
  }
  if (ctx.timeRange) {
    parts.push(`timeRange="${ctx.timeRange}"`);
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

  return ctx;
}

export function parseInlineAttr(content: string, key: string) {
  const match = new RegExp(`${key}="([^"]+?)"`).exec(content);
  if (!match) return null;
  return match[1];
}
