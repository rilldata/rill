import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { maybeGetMetricsResolverQueryFromMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query.ts";
import { get } from "svelte/store";

export async function load({ parent }) {
  const { message } = await parent();

  const query = maybeGetMetricsResolverQueryFromMessage(message);
  const rt = get(runtime);

  await openQuery({
    query,
    runtime: rt,
  });
}
