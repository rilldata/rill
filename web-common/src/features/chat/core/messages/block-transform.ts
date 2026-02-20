/**
 * Block Transformation
 *
 * Transforms raw API messages (V1Message) into UI blocks (Block).
 *
 * This module is the single source of truth for how messages become UI elements.
 * It handles:
 * - Routing messages to appropriate block types (text, thinking, chart, etc.)
 * - Filtering hidden messages (internal tools, results shown inline)
 * - Extracting feedback data and attaching it to assistant messages
 * - Correlating tool calls with their results
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { type ChartBlock } from "./chart/chart-block";
import { type FileDiffBlock } from "./file-diff/file-diff-block";
import {
  createTextBlock,
  type FeedbackData,
  type TextBlock,
} from "./text/text-block";
import {
  createThinkingBlock,
  type ThinkingBlock,
} from "./thinking/thinking-block";
import {
  getToolConfig,
  isHiddenTool,
  type ToolConfig,
  ToolGroupTypes,
} from "./tools/tool-registry";
import { shouldShowWorking, type WorkingBlock } from "./working/working-block";
import type { SimpleToolCall } from "@rilldata/web-common/features/chat/core/messages/simple-tool-call/simple-tool-call.ts";
import {
  createDevelopBlock,
  type DevelopBlock,
} from "@rilldata/web-common/features/chat/core/messages/develop/develop-block.ts";
import type { RestoreChangesBlock } from "@rilldata/web-common/features/chat/core/messages/restore/restore-block.ts";

// =============================================================================
// TYPES
// =============================================================================

export type Block =
  | TextBlock
  | ThinkingBlock
  | ChartBlock
  | FileDiffBlock
  | WorkingBlock
  | DevelopBlock
  | RestoreChangesBlock
  | SimpleToolCall;

export type {
  ChartBlock,
  FileDiffBlock,
  TextBlock,
  ThinkingBlock,
  WorkingBlock,
  SimpleToolCall,
};

// =============================================================================
// TRANSFORMATION
// =============================================================================

/**
 * Transforms raw chat messages into a list of UI blocks.
 */
export function transformToBlocks(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): Block[] {
  const blocks: Block[] = [];

  // Build lookup maps.
  // Feedback is pre-built (unlike other blocks) because it attaches to a different
  // message — the target assistant message appears before its feedback CALL/RESULT
  // messages in the array, so we need the data ready before we encounter the target.
  const resultMap = buildResultMessageMap(messages);
  const feedbackMap = buildFeedbackMap(messages);
  const groupedMessages = new Map<ToolGroupTypes, V1Message[]>();

  function flushThinking(isComplete: boolean): void {
    const thinkingMessages = groupedMessages.get(ToolGroupTypes.Thinking);
    if (!thinkingMessages || thinkingMessages.length === 0) {
      return;
    }
    groupedMessages.delete(ToolGroupTypes.Thinking);

    blocks.push(
      createThinkingBlock(
        thinkingMessages,
        resultMap,
        `thinking-${blocks.length}`,
        isComplete,
      ),
    );
  }

  function flushDevelop(): void {
    const developMessages = groupedMessages.get(ToolGroupTypes.Develop);
    if (!developMessages || developMessages.length === 0) {
      return;
    }
    groupedMessages.delete(ToolGroupTypes.Develop);

    const block = createDevelopBlock(
      developMessages,
      `develop-${blocks.length}`,
      resultMap,
    );
    if (block) blocks.push(block);
  }

  // Process each message
  for (const msg of messages) {
    const routing = getBlockRoute(msg);

    switch (routing.route) {
      case "text": {
        // Text blocks close any open blocks
        flushThinking(true);
        flushDevelop();
        // Attach feedback for assistant messages (router_agent results)
        const feedback =
          msg.role === "assistant" ? feedbackMap.get(msg.id!) : undefined;
        blocks.push(createTextBlock(msg, feedback));
        break;
      }

      case "group": {
        const groups = routing.config?.groups ?? ["thinking"];
        groups.forEach((group) => {
          if (groupedMessages.has(group)) {
            groupedMessages.get(group)!.push(msg);
          } else {
            groupedMessages.set(group, [msg]);
          }
        });

        break;
      }

      case "block": {
        flushThinking(true);
        const block = routing.config.createBlock?.(
          msg,
          resultMap.get(msg.id),
          messages,
        );
        if (block) {
          blocks.push(block);
        }
        break;
      }

      case "skip":
        break;
    }
  }

  // Flush remaining thinking messages
  flushThinking(!isStreaming && !isConversationLoading);
  // Flush remaining develop messages
  flushDevelop();

  // Add working indicator if AI is still processing
  if (shouldShowWorking(blocks, isStreaming, isConversationLoading)) {
    blocks.push({ type: "working", id: "working-indicator" });
  }

  return blocks;
}

// =============================================================================
// HELPERS
// =============================================================================

// ----- Routing -----

type BlockRoute =
  | { route: "text" }
  | { route: "group"; config?: ToolConfig }
  | { route: "block"; config: ToolConfig }
  | { route: "skip" };

/**
 * Determines where a message should be routed for rendering.
 */
function getBlockRoute(msg: V1Message): BlockRoute {
  // Visibility check
  if (shouldHideMessage(msg)) {
    return { route: "skip" };
  }

  // Router agent → text (main conversation)
  if (msg.tool === ToolName.ROUTER_AGENT) {
    return { route: "text" };
  }

  // Progress → thinking
  if (msg.type === MessageType.PROGRESS) {
    return { route: "group" };
  }

  // Tool calls → consult registry for block vs inline
  if (msg.type === MessageType.CALL) {
    const config = getToolConfig(msg.tool);
    return config.renderMode === "block"
      ? { route: "block", config }
      : { route: "group", config };
  }

  return { route: "skip" };
}

// ----- Visibility -----
//
// "Hidden" messages are still rendered, just in different UI locations:
// - Internal tools (analyst_agent, feedback_agent, etc.) → thinking blocks
// - Tool results → inside their parent tool call's collapsible UI
// - Feedback-related router_agent messages → feedback shown inline on target message
//
// Exception: router_agent results are the AI's text responses and ARE shown
// as main chat blocks (unless they're feedback-related).

function shouldHideMessage(msg: V1Message): boolean {
  // Internal tools → rendered in thinking blocks (see tool-registry.ts)
  if (msg.tool !== ToolName.ROUTER_AGENT && isHiddenTool(msg.tool)) {
    return true;
  }

  // Tool results → rendered inside parent tool call's UI
  if (msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT) {
    return true;
  }

  // Feedback goes through router_agent, producing both feedback_agent messages (hidden by
  // the registry above) AND router_agent wrapper messages. Without this check, those wrapper
  // messages would render as text blocks since router_agent is normally the main conversation.
  if (msg.tool === ToolName.ROUTER_AGENT && isFeedbackRouterMessage(msg)) {
    return true;
  }

  return false;
}

/**
 * Check if a router_agent message is feedback-related.
 * The backend marks these with agent: "feedback_agent" in the content.
 */
function isFeedbackRouterMessage(msg: V1Message): boolean {
  if (!msg.contentData) return false;
  try {
    const content = JSON.parse(msg.contentData);
    return content.agent === ToolName.FEEDBACK_AGENT;
  } catch {
    return false;
  }
}

// ----- Result Correlation -----

/**
 * Build a map from tool call message IDs to their result messages.
 */
function buildResultMessageMap(
  messages: V1Message[],
): Map<string | undefined, V1Message> {
  return new Map(
    messages
      .filter(
        (msg) =>
          msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT,
      )
      .map((msg) => [msg.parentId, msg]),
  );
}

// ----- Feedback Extraction -----

/** Structure of feedback call content stored in messages */
interface FeedbackCallContent {
  target_message_id: string;
  sentiment: "positive" | "negative";
  categories?: string[];
  comment?: string;
}

/**
 * Build a map of feedback data keyed by target message ID.
 * Scans for feedback_agent CALL and RESULT messages.
 */
function buildFeedbackMap(messages: V1Message[]): Map<string, FeedbackData> {
  const feedbackMap = new Map<string, FeedbackData>();

  for (const msg of messages) {
    const isFeedbackCall =
      msg.tool === ToolName.FEEDBACK_AGENT && msg.type === MessageType.CALL;
    if (!isFeedbackCall) continue;

    try {
      const content = JSON.parse(msg.contentData || "") as FeedbackCallContent;

      if (!content.target_message_id || !content.sentiment) continue;

      // Find the corresponding RESULT
      const resultMsg = messages.find(
        (m) =>
          m.tool === ToolName.FEEDBACK_AGENT &&
          m.type === MessageType.RESULT &&
          m.parentId === msg.id,
      );

      let response: string | null = null;
      if (resultMsg) {
        try {
          const resultContent = JSON.parse(resultMsg.contentData || "");
          response = resultContent.response || null;
        } catch {
          // Skip malformed result
        }
      }

      feedbackMap.set(content.target_message_id, {
        sentiment: content.sentiment,
        response,
        isPending: !resultMsg,
      });
    } catch {
      // Skip malformed feedback messages
    }
  }

  return feedbackMap;
}
