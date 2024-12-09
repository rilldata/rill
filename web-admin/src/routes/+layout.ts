/* 
In svelte.config.js, the "adapter-static" option makes the application a single-page
app in production. Here, we are setting server-side rendering (SSR) to false to 
ensure the same single-page app behavior in development.
*/
export const ssr = false;

import { dev } from "$app/environment";
import {
  type V1OrganizationPermissions,
  type V1ProjectPermissions,
} from "@rilldata/web-admin/client";
import { redirectToLoginOrRequestAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { fetchOrganizationPermissions } from "@rilldata/web-admin/features/organizations/selectors";
import { fetchProjectDeploymentDetails } from "@rilldata/web-admin/features/projects/selectors";
import { initPosthog } from "@rilldata/web-common/lib/analytics/posthog";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { error, redirect, type Page } from "@sveltejs/kit";

export const load = async ({ params, url, route }) => {
  // Route params
  const { organization, project, token: routeToken } = params;
  const pageState = {
    url,
    route,
    params,
  } as Page;

  let searchParamToken: string | undefined;
  if (url.searchParams.has("token")) {
    searchParamToken = url.searchParams.get("token") ?? undefined;
  }
  const token = searchParamToken ?? routeToken;

  // Initialize analytics
  const shouldSendAnalytics = !import.meta.env.VITE_PLAYWRIGHT_TEST && !dev;
  if (shouldSendAnalytics) {
    const rillVersion = import.meta.env.RILL_UI_VERSION;
    const posthogSessionId = url.searchParams.get("ph_session_id") as
      | string
      | null;
    initPosthog(rillVersion, posthogSessionId);
    if (posthogSessionId) {
      // Remove the PostHog sessionID from the url
      url.searchParams.delete("ph_session_id");
      throw redirect(307, url.toString());
    }
  }

  // If no organization or project, return empty permissions
  if (!organization) {
    return {
      organizationPermissions: <V1OrganizationPermissions>{},
      projectPermissions: <V1ProjectPermissions>{},
    };
  }

  // Get organization permissions
  let organizationPermissions: V1OrganizationPermissions = {};
  if (organization && !token) {
    try {
      organizationPermissions =
        await fetchOrganizationPermissions(organization);
    } catch (e) {
      if (e.response?.status !== 403) {
        throw error(e.response.status, "Error fetching organization");
      }
    }
  }

  if (!project) {
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
      runtimeData.jwt?.token,
      runtimeData.jwt?.authContext,
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
