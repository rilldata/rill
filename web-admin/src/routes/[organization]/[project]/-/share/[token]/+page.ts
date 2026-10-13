import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { error, redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({
  params: { organization, project, token },
  url,
  parent,
}) => {
  // The root layout fetches the token's metadata and resolves the bearer token.
  const { publicURLLocked, tokenMetadata } = await parent();

  // While the token is locked, the share layout renders a password prompt instead of this page.
  if (publicURLLocked) return;

  // Public URLs specify the resource in the token's metadata
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
  if (tokenMetadata.state) {
    if (canvasResource) {
      // For canvas, state is already URL params, use them directly
      redirectUrl.search = tokenMetadata.state;
    } else {
      // For explore, state is proto-serialized, wrap in state param
      redirectUrl.search = new URLSearchParams({
        state: tokenMetadata.state,
      }).toString();
    }
  }

  throw redirect(307, redirectUrl.toString());
};
