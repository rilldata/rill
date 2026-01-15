import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params, url, parent }) => {
  // Only proceed once the runtime in parent is ready
  const parentData = await parent();

  // Get the organization and project from the URL
  const organization = params.organization;
  const project = params.project;
  const runtime = parentData.runtime;

  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ url, organization, project, runtime });
};
