import { getLocalRuntimeClient } from "../../../../../../../../../lib/local-runtime-config";
import { maybeGetMetricsResolverQueryFromMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";

export async function load({ parent }) {
  const { message } = await parent();

  const query = maybeGetMetricsResolverQueryFromMessage(message);
  const client = getLocalRuntimeClient();

  await openQuery({
    query,
    runtime: { host: client.host, instanceId: client.instanceId },
  });
}
