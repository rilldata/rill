import {
  DeployingDashboardUrlParam,
  PreCommitShaUrlParam,
} from "@rilldata/web-common/features/project/deploy/utils";

// /-/invite for the first deploy; /-/deploying when prod already exists.
// If the user was editing a dashboard, pass its name through so the
// destination page can route back once reconciliation completes.
// `preCommitSha` is the prod ProjectParser commit SHA captured at click
// time; the deploying page waits for prod to advance past it before
// redirecting, so the user doesn't land on stale content.
export function buildPostMergeUrl({
  organization,
  project,
  pathname,
  hadProdDeployment,
  preCommitSha,
}: {
  organization: string;
  project: string;
  pathname: string;
  hadProdDeployment: boolean;
  preCommitSha?: string;
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
    url.searchParams.set(DeployingDashboardUrlParam, dashboard);
  }
  if (preCommitSha) {
    url.searchParams.set(PreCommitShaUrlParam, preCommitSha);
  }
  return `${url.pathname}${url.search}`;
}
