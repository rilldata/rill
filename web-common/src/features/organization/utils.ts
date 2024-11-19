// TODO: Find a better solution and get a url from backend.
//       We should add an endpoint to get frontendUrl from the urls.go util on cloud.
export function buildPlanUpgradeUrl(
  org: string,
  adminUrl: string,
  isEmptyOrg: boolean,
  currentUrl: URL,
) {
  let cloudUrl = adminUrl.replace("admin.rilldata", "ui.rilldata");
  // hack for dev env
  if (cloudUrl === "http://localhost:9090") {
    cloudUrl = "http://localhost:3000";
  }

  const url = new URL(cloudUrl);
  if (isEmptyOrg) {
    // Empty org wont have billing related options so show the general setting page in the background
    url.pathname = `/${org}/-/settings`;
  } else {
    url.pathname = `/${org}/-/settings/billing`;
  }
  url.searchParams.set("upgrade", "true");
  const redirectUrl = new URL(currentUrl);
  // set the org to avoid showing the org selector again
  redirectUrl.searchParams.set("org", org);
  url.searchParams.set("redirect", redirectUrl.toString());
  return url.toString();
}
