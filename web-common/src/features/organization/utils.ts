// TODO: Find a better solution and get a url from backend.
//       We should add an endpoint to get frontendUrl from the urls.go util on cloud.
export function buildPlanUpgradeUrl(
  org: string,
  adminUrl: string,
  curUrl: URL,
) {
  let cloudUrl = adminUrl.replace("admin.rilldata", "ui.rilldata");
  // hack for dev env
  if (cloudUrl === "http://localhost:9090") {
    cloudUrl = "http://localhost:3000";
  }

  const url = new URL(cloudUrl);
  // TODO: this should be to general settings page
  url.pathname = `/${org}/-/settings/billing`;
  url.searchParams.set("upgrade", "true");
  // set the org to avoid showing the org selector
  const newCurUrl = new URL(curUrl);
  newCurUrl.searchParams.set("org", org);
  url.searchParams.set("redirect", newCurUrl.toString());
  return url.toString();
}
