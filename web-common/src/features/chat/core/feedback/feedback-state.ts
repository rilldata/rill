/**
 * State container for user feedback on AI messages.
 *
 * This class manages the reactive state for feedback UI (sentiments, responses, loading).
 * It does NOT handle submission - that's owned by the Conversation class, which updates
 * this state after submitting feedback to the server.
 *
 * Why separate from Conversation?
 * - Keeps Conversation focused on message flow and streaming
 * - Encapsulates feedback-specific state and hydration logic
 * - Provides a clean interface for UI components to read feedback state
 *
 * Boundary:
 * - FeedbackState: owns state (reads/writes to stores)
 * - Conversation: owns behavior (submission, streaming, updating this state)
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { writable, type Readable } from "svelte/store";
import { MessageType, ToolName } from "../types";
import type { FeedbackSentiment } from "./types";

/** Tool name used for feedback messages (matches backend UserFeedbackToolName) */
export const USER_FEEDBACK_TOOL = "user_feedback";

/** Structure of feedback call content stored in messages */
export interface FeedbackCallContent {
  target_message_id: string;
  sentiment: FeedbackSentiment;
  categories?: string[];
  comment?: string;
}

export class FeedbackState {
  // Sentiment for each rated message (positive/negative)
  private readonly _sentiments = writable(new Map<string, FeedbackSentiment>());
  public readonly sentiments: Readable<Map<string, FeedbackSentiment>> =
    this._sentiments;

  // AI-generated response text for each feedback submission (keyed by target message ID)
  private readonly _responses = writable(new Map<string, string>());
  public readonly responses: Readable<Map<string, string>> = this._responses;

  // Message ID currently being processed (for loading indicator)
  private readonly _pendingMessageId = writable<string | null>(null);
  public readonly pendingMessageId: Readable<string | null> =
    this._pendingMessageId;

  /**
   * Restore feedback state from conversation messages.
   * Called when conversation data loads to reflect persisted feedback.
   */
  hydrateFromMessages(messages: V1Message[]): void {
    const sentimentMap = new Map<string, FeedbackSentiment>();
    const responseMap = new Map<string, string>();

    // Build message map for parent lookups
    const messageMap = new Map(
      messages.filter((m) => m.id).map((m) => [m.id!, m]),
    );

    for (const msg of messages) {
      // Extract sentiments from user_feedback call messages
      if (msg.tool === USER_FEEDBACK_TOOL && msg.type === MessageType.CALL) {
        if (!msg.contentData) continue;
        try {
          const content = JSON.parse(msg.contentData) as FeedbackCallContent;
          if (content.target_message_id && content.sentiment) {
            sentimentMap.set(content.target_message_id, content.sentiment);
          }
        } catch {
          // Skip malformed feedback messages
        }
      }

      // Extract feedback responses from user_feedback result messages
      const feedbackResponse = this.extractFeedbackResponse(msg, messageMap);
      if (feedbackResponse) {
        responseMap.set(
          feedbackResponse.targetMessageId,
          feedbackResponse.response,
        );
      }
    }

    if (sentimentMap.size > 0) {
      this._sentiments.set(sentimentMap);
    }
    if (responseMap.size > 0) {
      this._responses.set(responseMap);
    }
  }

  // ----- State setters (called by Conversation) -----

  setSentiment(messageId: string, sentiment: FeedbackSentiment): void {
    this._sentiments.update((map) => {
      map.set(messageId, sentiment);
      return new Map(map);
    });
  }

  setResponse(messageId: string, response: string): void {
    this._responses.update((map) => {
      map.set(messageId, response);
      return new Map(map);
    });
  }

  setPending(messageId: string | null): void {
    this._pendingMessageId.set(messageId);
  }

  // ----- Private helpers -----

  /**
   * Extract feedback response data from a user_feedback result message.
   * Uses parent relationship to get the target_message_id from the call.
   */
  private extractFeedbackResponse(
    msg: V1Message,
    messageMap: Map<string, V1Message>,
  ): { response: string; targetMessageId: string } | null {
    if (
      msg.tool !== ToolName.USER_FEEDBACK ||
      msg.type !== MessageType.RESULT
    ) {
      return null;
    }
    try {
      const parsed = JSON.parse(msg.contentData || "");
      const response = parsed.response || "";

      // Look up the parent call to get target_message_id
      if (!msg.parentId) return null;
      const parentCall = messageMap.get(msg.parentId);
      if (!parentCall) return null;

      const parentParsed = JSON.parse(parentCall.contentData || "");
      const targetMessageId = parentParsed.target_message_id;
      if (!targetMessageId) return null;

      return { response, targetMessageId };
    } catch {
      return null;
    }
  }
}
