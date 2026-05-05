// /-/invite for the first deploy; /-/deploying when prod already exists.
// If the user was editing a dashboard, pass its name through so the
// destination page can route back once reconciliation completes.
export function buildPostMergeUrl({
  organization,
  project,
  pathname,
  hadProdDeployment,
}: {
  organization: string;
  project: string;
  pathname: string;
  hadProdDeployment: boolean;
}): string {
  const dashboard = pathname.match(
    /\/-\/edit\/(?:explore|canvas)\/([^/?#]+)/,
  )?.[1];

  const path = hadProdDeployment ? "/-/deploying" : "/-/invite";
  const url = new URL(
    `/${organization}/${project}${path}`,
    window.location.origin,
  );
  if (dashboard) {
    url.searchParams.set("deploying_dashboard", dashboard);
  }
  return `${url.pathname}${url.search}`;
}
