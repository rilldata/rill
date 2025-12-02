import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";
import type { Readable } from "svelte/store";
import {
  getActiveExploreContext,
  getActiveFileContext,
} from "@rilldata/web-common/features/chat/core/context/agent-contexts.ts";

export type ChatConfig = {
  agent: string;
  additionalContextStore: Readable<Partial<RuntimeServiceCompleteBody>>;
  emptyChatLabel: string;
  placeholder: string;
  enableMention: boolean;
};

export const dashboardChatConfig = <ChatConfig>{
  agent: ToolName.ANALYST_AGENT,
  additionalContextStore: getActiveExploreContext(),
  emptyChatLabel: "Happy to help explore your data",
  placeholder: "Ask about your data...",
  enableMention: true,
};

export const developerChatConfig = <ChatConfig>{
  agent: ToolName.DEVELOPER_AGENT,
  additionalContextStore: getActiveFileContext(),
  emptyChatLabel: "Happy to assist you make changes to the project",
  placeholder: "What change can I help you make...",
  enableMention: false,
};
