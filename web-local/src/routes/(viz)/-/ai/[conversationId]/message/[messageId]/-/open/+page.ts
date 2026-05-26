import { getLocalRuntimeClient } from "../../../../../../../../../lib/runtime-client";
import {
  getResolvedTimeRangesFromMessage,
  maybeGetMetricsResolverQueryFromMessage,
} from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";

export async function load({ parent }) {
  const { message, result } = await parent();

  const query = maybeGetMetricsResolverQueryFromMessage(message);
  const resolvedTimeRanges = result
    ? getResolvedTimeRangesFromMessage(result)
    : [];
  const client = getLocalRuntimeClient();

  await openQuery({
    query,
    resolvedTimeRanges,
    client,
  });
}
