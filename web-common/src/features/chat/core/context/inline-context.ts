import {
  ChatContextEntryType,
  ContextTypeData,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export const INLINE_CHAT_CONTEXT_TAG = "inline";

export type InlineChatContext = {
  type: ChatContextEntryType;
  label?: string;
  // Hierarchy of values.
  // EG, for ChatContextEntryType.DimensionValue, [metricsViewName, dimensionName, ...dimensionValues]
  values: string[];
};

export type InlineChatContextMetadata = Record<string, MetricsViewMetadata>;
export type MetricsViewMetadata = {
  metricsViewSpec: V1MetricsViewSpec;
  measures: Record<string, MetricsViewSpecMeasure>;
  dimensions: Record<string, MetricsViewSpecDimension>;
};

export function inlineChatContextsAreEqual(
  ctx1: InlineChatContext,
  ctx2: InlineChatContext,
) {
  if (ctx1.type !== ctx2.type) return false;
  if (ctx1.values.length !== ctx2.values.length) return false;
  return ctx1.values.every((v, i) => v === ctx2.values[i]);
}

export function convertContextToInlinePrompt(ctx: InlineChatContext) {
  const parts: string[] = [];

  switch (ctx.type) {
    case ChatContextEntryType.MetricsView:
      parts.push(`metrics_view="${ctx.values[0]}"`);
      break;

    case ChatContextEntryType.TimeRange:
      parts.push(`time_range="${ctx.values[0]}"`);
      break;

    case ChatContextEntryType.Measures:
      parts.push(`metrics_view="${ctx.values[0]}"`);
      parts.push(`measure="${ctx.values[1]}"`);
      break;

    case ChatContextEntryType.Dimensions:
      parts.push(`metrics_view="${ctx.values[0]}"`);
      parts.push(`dimension="${ctx.values[1]}"`);
      break;

    case ChatContextEntryType.DimensionValues:
      parts.push(`metrics_view="${ctx.values[0]}"`);
      parts.push(`dimension="${ctx.values[1]}"`);
      parts.push(...ctx.values.slice(2).map((v, i) => `value_${i}="${v}"`));
      break;
  }

  return `<${INLINE_CHAT_CONTEXT_TAG}>${parts.join(" ")}</${INLINE_CHAT_CONTEXT_TAG}>`;
}

const PARTS_REGEX = /(\w+?)="([^"]+?)"/g;

export function convertPromptValueToContext(
  contextValue: string,
): InlineChatContext | null {
  const parts = contextValue.matchAll(PARTS_REGEX);
  if (!parts) return null;

  const matchedKeys: string[] = [];
  const matchedValues: string[] = [];
  for (const [, key, value] of parts) {
    matchedKeys.push(key);
    matchedValues.push(value);
  }

  let type: ChatContextEntryType | null = null;
  if (matchedKeys[0] === "time_range") {
    type = ChatContextEntryType.TimeRange;
  } else if (matchedKeys[0] === "metrics_view") {
    type = ChatContextEntryType.MetricsView;

    if (matchedKeys[1] === "measure") {
      type = ChatContextEntryType.Measures;
    } else if (matchedKeys[1] === "dimension" && matchedKeys.length === 2) {
      type = ChatContextEntryType.Dimensions;
    } else if (matchedKeys[1] === "dimension" && matchedKeys.length > 2) {
      type = ChatContextEntryType.DimensionValues;
    }
  }

  if (!type) return null;

  return <InlineChatContext>{
    type,
    values: matchedValues,
  };
}

const ChatContextRegex = new RegExp(
  `<${INLINE_CHAT_CONTEXT_TAG}>(.*?)</${INLINE_CHAT_CONTEXT_TAG}>`,
  "gm",
);
export function convertPromptWithInlineContextToHTML(
  prompt: string,
  meta: InlineChatContextMetadata,
) {
  const lines = prompt.split("\n");
  const htmlLines = lines.map((line) =>
    line.replaceAll(ChatContextRegex, (raw, contextValue: string) => {
      const entry = convertPromptValueToContext(contextValue);
      if (!entry) return raw;

      const data = ContextTypeData[entry.type];
      if (!data) return raw;
      const label = data.getLabel(entry, meta);

      // TODO: once we support editing messages, embed other parts of context here.
      return `<span class="underline">${label}</span>`;
    }),
  );
  return htmlLines.join("<br>");
}
