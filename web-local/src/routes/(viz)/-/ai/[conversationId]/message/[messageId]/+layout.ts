import { getLocalRuntimeClient } from "../../../../../../../lib/local-runtime-config";
import { fetchMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ params: { conversationId, messageId } }) {
  const client = getLocalRuntimeClient();

  const message = await fetchMessage(
    { host: client.host, instanceId: client.instanceId },
    conversationId,
    messageId,
  );

  return {
    message,
  };
}
