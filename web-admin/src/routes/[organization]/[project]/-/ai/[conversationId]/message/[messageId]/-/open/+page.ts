import { maybeGetMetricsResolverQueryFromMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";

export async function load({ parent, params: { organization, project } }) {
  const { runtime, message } = await parent();

  const query = maybeGetMetricsResolverQueryFromMessage(message);

  await openQuery({
    query,
    runtime,
    organization,
    project,
  });
}
