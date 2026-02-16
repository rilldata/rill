import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { get } from "svelte/store";
import { fetchQueryForCall } from "@rilldata/web-common/features/chat/core/open-query-utils.ts";

export async function load({ params: { conversationId, callId } }) {
  const rt = get(runtime);

  const query = await fetchQueryForCall(rt.instanceId, conversationId, callId);

  await openQuery({
    query,
    runtime: rt,
  });
}
