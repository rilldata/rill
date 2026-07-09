import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({
  params: { organization, project, token },
  url,
}) => {
  // Public URLs specify the resource in the token's metadata
  const tokenResp = await fetchMagicAuthToken(token).catch((e) => {
    console.error(e);
    throw error(404, m.route_error_token_not_found());
  });

  const { token: tokenMetadata } = tokenResp;
  if (!tokenMetadata?.resources) {
    console.error("Token does not have any associated resources");
    throw error(404, m.route_error_token_resource_not_found());
  }

  // Check for explore resource
  const exploreResource = tokenMetadata.resources.find(
    (r) => r.type === ResourceKind.Explore,
  );

  // Check for canvas resource
  const canvasResource = tokenMetadata.resources.find(
    (r) => r.type === ResourceKind.Canvas,
  );

  if (!exploreResource && !canvasResource) {
    console.error("Token does not have an associated explore or canvas");
    throw error(404, m.route_error_dashboard_not_found());
  }

  // Determine which resource type to redirect to
  const resourceType = exploreResource ? "explore" : "canvas";
  const resourceName = exploreResource?.name || canvasResource?.name;

  const redirectUrl = new URL(
    `/${organization}/${project}/-/share/${token}/${resourceType}/${resourceName}`,
    url.origin,
  );

  // Get the initial state from the token
  if (tokenResp?.token?.state) {
    if (canvasResource) {
      // For canvas, state is already URL params, use them directly
      redirectUrl.search = tokenResp.token.state;
    } else {
      // For explore, state is proto-serialized, wrap in state param
      redirectUrl.search = new URLSearchParams({
        state: tokenResp.token.state,
      }).toString();
    }
  }

  throw redirect(307, redirectUrl.toString());
};
