import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { get } from "svelte/store";
import { fetchMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ params: { conversationId, messageId } }) {
  const rt = get(runtime);

  const message = await fetchMessage(rt, conversationId, messageId);

  return {
    message,
  };
}
