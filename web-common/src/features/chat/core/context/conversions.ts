import {
  type ChatContextEntry,
  ChatContextEntryType,
  type ContextMetadata,
  ContextTypeData,
} from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";

export const PROMPT_INLINE_CONTEXT_TAG = "inline";
export const CONTEXT_TYPE_ATTR = "data-context-type";
export const CONTEXT_VALUE_ATTR = "data-context-value";
export const CONTEXT_SUB_VALUE_ATTR = "data-context-sub-value";

const SEPARATOR = "\x01";

const ChatContextRegex = new RegExp(
  `<${PROMPT_INLINE_CONTEXT_TAG}>(.*?)</${PROMPT_INLINE_CONTEXT_TAG}>`,
  "gm",
);

export function convertContextToAttrs(ctx: ChatContextEntry) {
  return {
    [CONTEXT_TYPE_ATTR]: ctx.type,
    [CONTEXT_VALUE_ATTR]: ctx.value,
    ...(ctx.subValue ? { [CONTEXT_SUB_VALUE_ATTR]: ctx.subValue } : {}),
  };
}

export function convertContextToInlinePrompt(ctx: ChatContextEntry) {
  let inlineContext = `${ctx.type}${SEPARATOR}${ctx.value}`;
  if (ctx.subValue) inlineContext += `${SEPARATOR}${ctx.subValue}`;
  return `<${PROMPT_INLINE_CONTEXT_TAG}>${inlineContext}</${PROMPT_INLINE_CONTEXT_TAG}>`;
}

export function convertContextToHtml(prompt: string, meta: ContextMetadata) {
  const lines = prompt.split("\n");
  const htmlLines = lines.map((line) =>
    line.replaceAll(ChatContextRegex, (raw, contextValue: string) => {
      const entry = convertContextValueToEntry(contextValue, meta);
      if (!entry) return raw;

      const attrs = convertContextToAttrs(entry);
      const attrsStr = Object.entries(attrs)
        .map(([k, v]) => `${k}="${v}"`)
        .join(" ");

      return `<span ${attrsStr} class="underline">${entry.label}</span>`;
    }),
  );
  return htmlLines.join("<br>");
}

export function convertContextValueToEntry(
  contextValue: string,
  meta: ContextMetadata,
): ChatContextEntry | undefined {
  const [type, value, subValue = null] = contextValue.split(SEPARATOR) as [
    ChatContextEntryType,
    string,
    string | undefined,
  ];

  const data = ContextTypeData[type];
  if (!data) return undefined;

  const entry = {
    type,
    value,
    subValue,
    label: "",
  };
  entry.label = data.getLabel(entry, meta);

  return entry;
}
