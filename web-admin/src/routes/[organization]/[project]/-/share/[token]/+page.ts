import { fetchMagicAuthToken } from "@rilldata/web-admin/features/projects/selectors";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({
  params: { organization, project, token },
  url,
}) => {
  // Public URLs specify the resource in the token's metadata
  const tokenData = await fetchMagicAuthToken(token).catch((e) => {
    console.error(e);
    throw error(404, "Unable to find token");
  });

  const {
    token: { resourceName, resources },
  } = tokenData;
  if (!resourceName && !resources) {
    console.error("Token does not have an associated resource");
    throw error(404, "Unable to find resource");
  }

  const exploreName = resources[0].name || resourceName; // `resourceName` is here for backwards compatibility

  const redirectUrl = new URL(
    `/${organization}/${project}/-/share/${token}/explore/${exploreName}`,
    url.origin,
  );

  // Get the initial state from the token
  if (tokenData?.token?.state) {
    redirectUrl.search = new URLSearchParams({
      state: tokenData.token.state,
    }).toString();
  }

  throw redirect(307, redirectUrl.toString());
};
