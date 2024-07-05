/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import { adminServiceGetProject } from "@rilldata/web-admin/client/index.js";
import { getAdminServiceGetProjectQueryKey } from "@rilldata/web-admin/client/index.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { error } from "@sveltejs/kit";
import type { QueryFunction } from "@tanstack/svelte-query";

export const load = async ({ params }) => {
  const { organization, project } = params;

  if (!organization || !project) {
    return {
      projectPermissions: {},
    };
  }

  const queryKey = getAdminServiceGetProjectQueryKey(organization, project);

  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  > = ({ signal }) => adminServiceGetProject(organization, project, {}, signal);

  try {
    const response = await queryClient.fetchQuery({
      queryFn: queryFunction,
      queryKey,
    });

    const { projectPermissions } = response;

    return {
      projectPermissions,
    };
  } catch (e) {
    console.error(e);
    throw error(404, "Unable to find project");
  }
};
