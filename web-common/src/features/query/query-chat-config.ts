import {
  type ChatConfig,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types";
import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { readable, type Readable } from "svelte/store";

const emptyContext: Readable<Partial<RuntimeServiceCompleteBody>> = readable(
  {},
);

export function createQueryChatConfig(): ChatConfig {
  return {
    agent: ToolName.ANALYST_AGENT,
    additionalContextStoreGetter: () => emptyContext,
    emptyChatLabel: "Ask questions about your data",
    placeholder: "Ask a question about your data...",
    minChatHeight: "min-h-[4rem]",
  };
}
