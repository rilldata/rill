import { openQuery } from "@rilldata/web-common/features/explore-mappers/open-query";
import { createRuntimeClientFromLayout } from "@rilldata/web-admin/lib/runtime-client-utils";
import type { PageLoad } from "./$types";
import { getQueryFromUrl } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";

export const load: PageLoad = async ({ params, url, parent }) => {
  const { runtime } = await parent();
  const client = createRuntimeClientFromLayout(runtime!);

  const query = getQueryFromUrl(url);

  await openQuery({
    query,
    organization: params.organization,
    project: params.project,
    client,
  });
};
