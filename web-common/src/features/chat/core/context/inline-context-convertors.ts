import {
  InlineContextConfig,
  INLINE_CHAT_CONTEXT_TAG,
  type InlineContext,
  type InlineContextMetadata,
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

const ChatContextRegex = new RegExp(
  `<${INLINE_CHAT_CONTEXT_TAG}>(.*?)</${INLINE_CHAT_CONTEXT_TAG}>`,
  "gm",
);
export function convertPromptWithInlineContextToHTML(
  prompt: string,
  meta: InlineContextMetadata,
) {
  const lines = prompt.split("\n");
  const htmlLines = lines.map((line) =>
    line.replaceAll(ChatContextRegex, (raw, contextValue: string) => {
      const entry = convertPromptValueToContext(contextValue);
      if (!entry) return raw;

      const data = InlineContextConfig[entry.type];
      if (!data) return raw;
      const label = data.getLabel(entry, meta);

      // TODO: once we support editing messages, embed other parts of context here.
      return `<span class="underline">${label}</span>`;
    }),
  );
  return htmlLines.join("<br>");
}
