import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

export async function load({ url }) {
  const rt = get(runtime);

  // Get the JSON-encoded query parameters
  const queryJSON = url.searchParams.get("query");
  if (!queryJSON) {
    throw new Error("query parameter is required");
  }

  // Parse and validate the query with proper type safety
  let query: MetricsResolverQuery;
  try {
    query = JSON.parse(queryJSON) as MetricsResolverQuery;
  } catch (e) {
    throw new Error(`Invalid query: ${e.message}`);
  }

  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ query, runtime: rt });
}
