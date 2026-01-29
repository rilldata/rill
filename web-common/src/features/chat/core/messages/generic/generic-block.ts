import type { V1Message } from "@rilldata/web-common/runtime-client";

/**
 * Generic block that doesnt really have specific rendering.
 */
export type GenericBlock = {
  type: "generic-block";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
};

export function createGenericBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): GenericBlock | null {
  if (!resultMessage) return null;
  return {
    type: "generic-block",
    id: `generic-block-${message.id}`,
    message,
    resultMessage,
  };
}
