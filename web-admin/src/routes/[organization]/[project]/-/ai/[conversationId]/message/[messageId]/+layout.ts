import { fetchMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ params: { conversationId, messageId }, parent }) {
  const { runtime } = await parent();

  const message = await fetchMessage(runtime, conversationId, messageId);

  return {
    message,
  };
}
