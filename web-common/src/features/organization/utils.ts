// TODO: Find a better solution and get a url from backend.
//       We should add an endpoint to get frontendUrl from the urls.go util on cloud.
export function buildPlanUpgradeUrl(org: string, adminUrl: string) {
  let cloudUrl = adminUrl.replace("admin.rilldata", "ui.rilldata");
  // hack for dev env
  if (cloudUrl === "http://localhost:9090") {
    cloudUrl = "http://localhost:3000";
  }

  const url = new URL(cloudUrl);
  url.pathname = `/${org}/-/settings/billing`;
  url.searchParams.set("upgrade", "true");
  return url.toString();
}
