import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export async function load({ url }) {
  const rt = get(runtime);

  const query = getQueryFromUrl(url);

  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ query, runtime: rt });
}
