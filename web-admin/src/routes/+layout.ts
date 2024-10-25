/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import {
  adminServiceGetProject,
  getAdminServiceGetProjectQueryKey,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { checkUserAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error } from "@sveltejs/kit";
import type { QueryFunction, QueryKey } from "@tanstack/svelte-query";
import {
  adminServiceGetProjectWithBearerToken,
  getAdminServiceGetProjectWithBearerTokenQueryKey,
} from "../features/public-urls/get-project-with-bearer-token.js";

export const load = async ({ params, url }) => {
  const { organization, project, token: routeToken } = params;

  if (!organization || !project) {
    return {
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  let searchParamToken: string | undefined;
  if (url.searchParams.has("token")) {
    searchParamToken = url.searchParams.get("token");
  }

  const token = searchParamToken ?? routeToken;

  let queryKey: QueryKey;
  let queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  >;

  if (token) {
    queryKey = getAdminServiceGetProjectWithBearerTokenQueryKey(
      organization,
      project,
      token,
      {},
    );

    queryFn = ({ signal }) =>
      adminServiceGetProjectWithBearerToken(
        organization,
        project,
        token,
        {},
        signal,
      );
  } else {
    queryKey = getAdminServiceGetProjectQueryKey(organization, project);

    queryFn = ({ signal }) =>
      adminServiceGetProject(organization, project, {}, signal);
  }

  try {
    const response = await queryClient.fetchQuery({
      queryFn,
      queryKey,
    });

    const { projectPermissions } = response;

    return {
      projectPermissions,
    };
  } catch (e) {
    if (e.response?.status !== 403 || (await checkUserAccess())) {
      throw error(e.response.status, "Error fetching deployment");
    }
  }
};
