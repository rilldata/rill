import {
  InlineContextConfig,
  INLINE_CHAT_CONTEXT_TAG,
  type InlineContext,
  type InlineContextMetadata,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import ReadonlyInlineContext from "@rilldata/web-common/features/chat/core/context/ReadonlyInlineContext.svelte";
import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";

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

const OpeningTag = `<${INLINE_CHAT_CONTEXT_TAG}>`;
const ClosingTag = `</${INLINE_CHAT_CONTEXT_TAG}>`;

type TextOrComponent = {
  isSvelteComponent: boolean;
  text?: string;
  component?: any;
  props?: Record<string, any>;
};
export function convertPromptWithInlineContextToComponents(prompt: string) {
  const lines = prompt.split("\n");
  return lines.map((line) => {
    const lineComponents: TextOrComponent[] = [];

    let cursor = 0;

    while (cursor < line.length) {
      const contextIndex = line.indexOf(OpeningTag, cursor);
      if (contextIndex === -1) break;

      const closingIndex = line.indexOf(ClosingTag, contextIndex);
      if (closingIndex === -1) break;

      const ctx = convertPromptValueToContext(
        line.substring(contextIndex + OpeningTag.length, closingIndex),
      );
      if (!ctx) break;

      lineComponents.push({
        isSvelteComponent: false,
        text: line.substring(cursor, line.indexOf(OpeningTag, contextIndex)),
      });

      lineComponents.push({
        isSvelteComponent: true,
        component: ReadonlyInlineContext,
        props: {
          chatContext: ctx,
        },
      });
      cursor = closingIndex + ClosingTag.length;
    }

    if (cursor < line.length) {
      lineComponents.push({
        isSvelteComponent: false,
        text: line.substring(cursor),
      });
    }

    return lineComponents;
  });
}
