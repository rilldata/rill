import type { V1Message } from "@rilldata/web-common/runtime-client";

/**
 * Generic block that doesnt really have specific rendering.
 */
export type SimpleToolCall = {
  type: "simple-tool-call-block";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
};

export function createSimpleTooCall(
  message: V1Message,
  resultMessage: V1Message | undefined,
): SimpleToolCall | null {
  if (!resultMessage) return null;
  return {
    type: "simple-tool-call-block",
    id: `generic-block-${message.id}`,
    message,
    resultMessage,
  };
}
