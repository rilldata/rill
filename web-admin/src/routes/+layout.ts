/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import {
  type V1OrganizationPermissions,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import {
  redirectToLoginIfNotLoggedIn,
  redirectToLoginOrRequestAccess,
} from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { fetchOrganizationPermissions } from "@rilldata/web-admin/features/organizations/selectors";
import { fetchProjectDeploymentDetails } from "@rilldata/web-admin/features/projects/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, type Page } from "@sveltejs/kit";

export const load = async ({ params, url, route }) => {
  const { organization, project, token: routeToken } = params;
  const pageState = {
    url,
    route,
    params,
  } as Page;

  let searchParamToken: string | undefined;
  if (url.searchParams.has("token")) {
    searchParamToken = url.searchParams.get("token");
  }
  const token = searchParamToken ?? routeToken;

  let organizationPermissions: V1OrganizationPermissions = {};
  if (organization && !token) {
    try {
      organizationPermissions =
        await fetchOrganizationPermissions(organization);
    } catch (e) {
      if (e.response?.status !== 403) {
        throw error(e.response.status, "Error fetching organization");
      }
      // Use without access to anything withing the org will hit this, so redirect to access page here.
      const didRedirect = await redirectToLoginIfNotLoggedIn();
      if (!didRedirect) {
        return {
          organizationPermissions,
          projectPermissions: <V1ProjectPermissions>{},
        };
      }
    }
  }

  if (!organization || !project) {
    return {
      organizationPermissions,
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  try {
    const {
      projectPermissions,
      project: proj,
      runtime: runtimeData,
    } = await fetchProjectDeploymentDetails(organization, project, token);

    await runtime.setRuntime(
      queryClient,
      fixLocalhostRuntimePort(runtimeData.host),
      runtimeData.instanceId,
      runtimeData.jwt.token,
      runtimeData.jwt.authContext,
    );

    return {
      organizationPermissions,
      projectPermissions,
      project: proj,
      runtime: runtimeData,
    };
  } catch (e) {
    if (e.response?.status !== 403) {
      throw error(e.response.status, "Error fetching deployment");
    }
    const didRedirect = await redirectToLoginOrRequestAccess(pageState);
    if (!didRedirect) {
      throw error(e.response.status, "Error fetching organization");
    }
  }
};
