import type { V1Message } from "@rilldata/web-common/runtime-client";

/**
 * Text block representation.
 * Contains a single text message (user or assistant).
 */
export type TextBlock = {
  type: "text";
  id: string;
  message: V1Message;
};

/**
 * Creates a text block from a message.
 */
export function createTextBlock(message: V1Message): TextBlock {
  return {
    type: "text",
    id: message.id!,
    message,
  };
}
