import {
  TargetDashboardUrlParam,
  PreCommitShaUrlParam,
} from "@rilldata/web-common/features/project/deploy/utils";
import type { Page } from "@sveltejs/kit";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { get } from "svelte/store";

// /-/invite for the first deploy; /-/deploying when prod already exists.
// If the user was editing a dashboard, pass its name through so the
// destination page can route back once reconciliation completes.
// `preCommitSha` is the prod ProjectParser commit SHA captured at click
// time; the deploying page waits for prod to advance past it before
// redirecting, so the user doesn't land on stale content.
export function buildPostMergeUrl({
  organization,
  project,
  page,
  hadProdDeployment,
  preCommitSha,
}: {
  organization: string;
  project: string;
  page: Page;
  hadProdDeployment: boolean;
  preCommitSha?: string;
}): string {
  const path = hadProdDeployment ? "/-/deploying" : "/-/invite";
  const url = new URL(
    `/${organization}/${project}${path}`,
    window.location.origin,
  );

  const dashboard = getDashboardFromUrl(page);
  if (dashboard) {
    url.searchParams.set(TargetDashboardUrlParam, dashboard);
  }

  if (preCommitSha) {
    url.searchParams.set(PreCommitShaUrlParam, preCommitSha);
  }
  return `${url.pathname}${url.search}`;
}

function getDashboardFromUrl(page: Page) {
  if (isEditRoute(page)) {
    const filePath = page.url.pathname.replace(/.*\/files/, "");
    const curFile = fileArtifacts.getFileArtifact(filePath);
    const curResource = get(curFile.resourceName);
    const isDashboardResource =
      curResource?.kind === ResourceKind.Explore ||
      curResource?.kind === ResourceKind.Canvas;
    return isDashboardResource ? curResource.name : undefined;
  }

  if (isDashboardVizRoute(page)) {
    return page.params.name;
  }
}

function isDashboardVizRoute(page: Page) {
  return (
    page.route.id === "/[organization]/[project]/-/edit/(viz)/explore/[name]" ||
    page.route.id === "/[organization]/[project]/-/edit/(viz)/canvas/[name]"
  );
}

function isEditRoute(page: Page) {
  return page.route.id?.startsWith(
    "/[organization]/[project]/-/edit/(workspace)/files",
  );
}
