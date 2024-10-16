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
  const { organization, project, token } = params;
  let effectiveToken = token;

  if (!organization || !project) {
    return {
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  if (url.searchParams.has("token")) {
    effectiveToken = url.searchParams.get("token");
  }

  let queryKey: QueryKey;
  let queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  >;

  console.log(organization, project, effectiveToken);
  if (effectiveToken) {
    queryKey = getAdminServiceGetProjectWithBearerTokenQueryKey(
      organization,
      project,
      effectiveToken,
      {},
    );

    queryFn = ({ signal }) =>
      adminServiceGetProjectWithBearerToken(
        organization,
        project,
        effectiveToken,
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
