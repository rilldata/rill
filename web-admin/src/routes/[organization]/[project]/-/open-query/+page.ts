import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params, url, parent }) => {
  // Only proceed once the runtime in parent is ready
  await parent();

  // Get the organization and project from the URL
  const organization = params.organization;
  const project = params.project;

  // Open the query (this'll redirect to the relevant Explore page)
  await openQuery({ url, organization, project });
};
