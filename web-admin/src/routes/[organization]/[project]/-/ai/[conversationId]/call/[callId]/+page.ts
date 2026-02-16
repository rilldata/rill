import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { fetchQueryForCall } from "@rilldata/web-common/features/chat/core/open-query-utils.ts";

export async function load({
  params: { organization, project, conversationId, callId },
  parent,
}) {
  const { runtime } = await parent();

  const query = await fetchQueryForCall(
    runtime.instanceId,
    conversationId,
    callId,
  );

  await openQuery({
    query,
    runtime,
    organization,
    project,
  });
}
