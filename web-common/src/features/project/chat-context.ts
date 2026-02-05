import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import { readable } from "svelte/store";

export const projectChat = {
  agent: ToolName.ANALYST_AGENT,
  additionalContextStoreGetter: () => readable({}),
  emptyChatLabel: "Happy to help explore your data",
  placeholder:
    "Type a question, or press @ to insert a metric, dimension, or measure.",
  minChatHeight: "min-h-[2.5rem]",
} satisfies ChatConfig;
