import { redirect } from "@sveltejs/kit";

// Unsubscribe links used to point here, so links in emails that were already sent are redirected to the current route.
// The current route takes the org and project as query params, since the unsubscribe token has no access to them.
export const load = ({ params, url }) => {
  const searchParams = new URLSearchParams(url.searchParams);
  searchParams.set("org", params.organization);
  searchParams.set("project", params.project);
  throw redirect(
    307,
    `/-/unsubscribe/alerts/${encodeURIComponent(params.alert)}?${searchParams.toString()}`,
  );
};
