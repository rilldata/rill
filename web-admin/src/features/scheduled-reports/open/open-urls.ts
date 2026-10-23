import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import { stripInternalReportParams } from "@rilldata/web-common/features/scheduled-reports/utils";
import type {
  V1AlertSpec,
  V1ReportSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Builds the paths of the dashboards that report and alert open links land on.
 * Rill Cloud and embeds serve dashboards from different routes.
 */
export type DashboardPathBuilder = {
  explore: (name: string) => string;
  canvas: (name: string) => string;
};

/**
 * Dashboard paths in Rill Cloud.
 * When opened with a magic token, dashboards are served from the token's share route.
 */
export function getCloudDashboardPaths(
  organization: string,
  project: string,
  token?: string | null,
): DashboardPathBuilder {
  const base = token
    ? `/${organization}/${project}/-/share/${token}`
    : `/${organization}/${project}`;
  return {
    explore: (name) => `${base}/explore/${name}`,
    canvas: (name) => `${base}/canvas/${name}`,
  };
}

/**
 * Dashboard paths in an embed.
 */
export const EmbedDashboardPaths: DashboardPathBuilder = {
  explore: (name) => `/-/embed/explore/${name}`,
  canvas: (name) => `/-/embed/canvas/${name}`,
};

/**
 * Returns the URL that opens a canvas report, with the canvas state captured at scheduling time applied.
 * Returns undefined if the report isn't for a canvas.
 */
export function getCanvasReportOpenUrl(
  reportSpec: V1ReportSpec,
  paths: DashboardPathBuilder,
): string | undefined {
  const canvasName = reportSpec.annotations?.["canvas"];
  if (!canvasName) return undefined;

  const path = paths.canvas(canvasName);
  const search = stripInternalReportParams(
    new URLSearchParams(reportSpec.annotations?.web_open_state ?? ""),
  ).toString();
  return search ? `${path}?${search}` : path;
}

/**
 * Returns the explore that a report or alert opens.
 */
export function getOpenExploreName(spec: V1ReportSpec | V1AlertSpec): string {
  return (
    spec.annotations?.["explore"] ??
    getExploreName(spec.annotations?.web_open_path ?? "") // backwards compatibility
  );
}

/**
 * Returns the query that a report or alert runs, which is mapped to the dashboard state when opened.
 */
export function getOpenQuery(spec: V1ReportSpec | V1AlertSpec): {
  queryName: string;
  queryArgsJson: string;
} {
  return {
    queryName:
      (spec.resolverProperties?.query_name as string | undefined) ??
      spec.queryName ??
      "",
    queryArgsJson:
      (spec.resolverProperties?.query_args_json as string | undefined) ??
      spec.queryArgsJson ??
      "",
  };
}
