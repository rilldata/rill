import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";
import type { Readable } from "svelte/store";
import {
  getActiveExploreContext,
  getActiveFileContext,
} from "@rilldata/web-common/features/chat/core/context/agent-contexts.ts";

export type ChatInputConfig = {
  agent: string;
  additionalContextStore: Readable<Partial<RuntimeServiceCompleteBody>>;
  placeholder: string;
  enableMention: boolean;
};

export const dashboardChatConfig = <ChatInputConfig>{
  agent: ToolName.ANALYST_AGENT,
  additionalContextStore: getActiveExploreContext(),
  placeholder: "Ask about your data...",
  enableMention: true,
};

export const developerChatConfig = <ChatInputConfig>{
  agent: ToolName.DEVELOPER_AGENT,
  additionalContextStore: getActiveFileContext(),
  placeholder: "What change can I help you make...",
  enableMention: false,
};
