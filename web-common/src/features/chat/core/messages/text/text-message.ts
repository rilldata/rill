import type { V1Message } from "@rilldata/web-common/runtime-client";

/**
 * Text message representation.
 * Contains a single text message (user or assistant).
 */
export type TextMessage = {
  type: "text";
  id: string;
  message: V1Message;
};

/**
 * Creates a text message block from a message.
 */
export function createTextMessage(
  message: V1Message,
  fallbackId: string,
): TextMessage {
  return {
    type: "text",
    id: message.id || fallbackId,
    message,
  };
}
