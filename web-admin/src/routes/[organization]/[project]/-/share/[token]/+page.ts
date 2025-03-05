import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({
  params: { organization, project, token },
  url,
}) => {
  // Public URLs specify the resource in the token's metadata
  const tokenResp = await fetchMagicAuthToken(token).catch((e) => {
    console.error(e);
    throw error(404, "Unable to find token");
  });

  const { token: tokenMetadata } = tokenResp;
  if (!tokenMetadata?.resources) {
    console.error("Token does not have any associated resources");
    throw error(404, "Unable to find the token's resource");
  }

  const exploreName = tokenMetadata.resources.find(
    (r) => r.type === ResourceKind.Explore,
  ); // Assumes only one explore per token

  if (!exploreName) {
    console.error("Token does not have an associated explore");
    throw error(404, "Unable to find an explore");
  }

  const redirectUrl = new URL(
    `/${organization}/${project}/-/share/${token}/explore/${exploreName.name}`,
    url.origin,
  );

  // Get the initial state from the token
  if (tokenResp?.token?.state) {
    redirectUrl.search = new URLSearchParams({
      state: tokenResp.token.state,
    }).toString();
  }

  throw redirect(307, redirectUrl.toString());
};
