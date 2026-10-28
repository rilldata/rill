import { InlineContextConfig } from "@rilldata/web-common/features/chat/core/context/config.ts";
import {
  convertPromptValueToContext,
  INLINE_CHAT_CONTEXT_TAG,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { InlineContextMetadata } from "@rilldata/web-common/features/chat/core/context/metadata.ts";

const CHAT_REFERENCE_REGEX = new RegExp(
  `<${INLINE_CHAT_CONTEXT_TAG}\\b([\\s\\S]*?)</${INLINE_CHAT_CONTEXT_TAG}>`,
  "g",
);
// Older titles were truncated to 47 bytes followed by "..." without regard for references,
// so they can end in a partial reference, even one cut within its tag name (e.g. `<chat-...`).
const TAG_NAME_PREFIXES = Array.from(INLINE_CHAT_CONTEXT_TAG, (_, i) =>
  INLINE_CHAT_CONTEXT_TAG.slice(0, i),
).join("|");
const PARTIAL_CHAT_REFERENCE_REGEX = new RegExp(
  `<(?:${INLINE_CHAT_CONTEXT_TAG}\\b[\\s\\S]*|(?:${TAG_NAME_PREFIXES})\\.\\.\\.)$`,
);

/**
 * Converts a conversation title to plain text by replacing the chat references in it with their labels.
 * The backend keeps references in titles as-is so that labels are always up-to-date with the project.
 */
export function formatConversationTitle(
  title: string,
  meta: InlineContextMetadata,
) {
  const text = title.replace(CHAT_REFERENCE_REGEX, (_, attrs: string) => {
    const ctx = convertPromptValueToContext(attrs);
    if (!ctx) return "";
    const label = InlineContextConfig[ctx.type].getLabel(ctx, meta);
    return ctx.type === InlineContextType.Skill ? `/${label}` : label;
  });

  const withoutPartial = text.replace(PARTIAL_CHAT_REFERENCE_REGEX, "").trim();
  if (withoutPartial !== text.trim() && withoutPartial) {
    return `${withoutPartial}...`;
  }
  return withoutPartial;
}

/**
 * Removes a trailing partial chat reference left by the old title truncation,
 * so that the title can be loaded into the chat editor without leaking broken markup.
 */
export function stripPartialChatReference(title: string) {
  // The regex matches from the first opening tag, so only apply it after the last complete reference.
  const closingTag = `</${INLINE_CHAT_CONTEXT_TAG}>`;
  const lastClosingTag = title.lastIndexOf(closingTag);
  const tailStart =
    lastClosingTag === -1 ? 0 : lastClosingTag + closingTag.length;
  const tail = title.slice(tailStart).replace(PARTIAL_CHAT_REFERENCE_REGEX, "");
  return (title.slice(0, tailStart) + tail).trim();
}
