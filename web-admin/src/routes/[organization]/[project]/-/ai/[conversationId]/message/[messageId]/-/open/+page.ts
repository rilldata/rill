import { maybeGetMetricsResolverQueryFromMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { createRuntimeClientFromLayout } from "@rilldata/web-admin/lib/runtime-client-utils";

export async function load({ parent, params: { organization, project } }) {
  const { runtime, message } = await parent();
  const client = createRuntimeClientFromLayout(runtime);

  const query = maybeGetMetricsResolverQueryFromMessage(message);

  await openQuery({
    query,
    client,
    organization,
    project,
  });
}
