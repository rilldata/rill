import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import * as m from "@rilldata/web-common/paraglide/messages.js";
import { readable } from "svelte/store";

export const projectChat = {
  agent: ToolName.ANALYST_AGENT,
  additionalContextStoreGetter: () => readable({}),
  get emptyChatLabel() {
    return m.chat_empty_label();
  },
  get placeholder() {
    return m.chat_placeholder_analyst();
  },
  minChatHeight: "min-h-[2.5rem]",
} satisfies ChatConfig;
