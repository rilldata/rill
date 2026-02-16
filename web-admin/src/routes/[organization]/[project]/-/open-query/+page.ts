import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import type { PageLoad } from "./$types";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

export const load: PageLoad = async ({ params, url, parent }) => {
  // Only proceed once the runtime in parent is ready
  const parentData = await parent();

  // Get the organization and project from the URL
  const organization = params.organization;
  const project = params.project;
  const runtime = parentData.runtime;

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
  await openQuery({ query, organization, project, runtime });
};
