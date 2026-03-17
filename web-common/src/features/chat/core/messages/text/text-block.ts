import type { V1Message } from "@rilldata/web-common/runtime-client";
import type { FeedbackSentiment } from "../../feedback/feedback-categories";

/**
 * Text block representation.
 * Contains a single text message (user or assistant).
 */
export type TextBlock = {
  type: "text";
  id: string;
  message: V1Message;
  feedback?: FeedbackData;
};

/**
 * Feedback data attached to assistant messages.
 * Derived from feedback_agent CALL/RESULT messages during block transformation.
 */
export interface FeedbackData {
  sentiment: FeedbackSentiment;
  response: string | null;
  isPending: boolean;
}

/**
 * Creates a text block from a message.
 */
export function createTextBlock(
  message: V1Message,
  feedback?: FeedbackData,
): TextBlock {
  return {
    type: "text",
    id: message.id!,
    message,
    feedback,
  };
}
