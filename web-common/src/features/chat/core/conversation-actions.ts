/**
 * Mutations for managing conversations from the conversation list and the chat header.
 */
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { V1GetConversationResponse } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getRuntimeServiceGetConversationQueryKey,
  runtimeServiceDeleteConversation,
  runtimeServiceUpdateConversation,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { invalidateConversationsList } from "./utils";

export async function renameConversation(
  client: RuntimeClient,
  conversationId: string,
  title: string,
) {
  await runtimeServiceUpdateConversation(client, { conversationId, title });

  // The conversation query is managed manually during streaming, so patch the title instead of refetching it.
  queryClient.setQueryData<V1GetConversationResponse>(
    getRuntimeServiceGetConversationQueryKey(client.instanceId, {
      conversationId,
    }),
    (data) =>
      data?.conversation
        ? { ...data, conversation: { ...data.conversation, title } }
        : data,
  );
  await invalidateConversationsList(client.instanceId);
}

export async function deleteConversation(
  client: RuntimeClient,
  conversationId: string,
) {
  await runtimeServiceDeleteConversation(client, { conversationId });
  await invalidateConversationsList(client.instanceId);
}
