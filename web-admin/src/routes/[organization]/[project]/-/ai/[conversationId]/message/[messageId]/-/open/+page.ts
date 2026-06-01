import {
  getResolvedTimeRangesFromMessage,
  maybeGetMetricsResolverQueryFromMessage,
} from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client";

export async function load({ parent, params: { organization, project } }) {
  const { runtime, message, result } = await parent();
  const client = getCloudRuntimeClient(runtime);

  const query = maybeGetMetricsResolverQueryFromMessage(message);
  const resolvedTimeRanges = result
    ? getResolvedTimeRangesFromMessage(result)
    : [];

  await openQuery({
    mapArgs: { query, ...resolvedTimeRanges },
    client,
    organization,
    project,
  });
}
