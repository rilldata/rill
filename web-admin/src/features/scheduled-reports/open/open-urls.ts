import {
  type DashboardPathContext,
  getDashboardPath,
  getExploreName,
} from "@rilldata/web-common/features/explore-mappers/utils";
import { stripInternalReportParams } from "@rilldata/web-common/features/scheduled-reports/utils";
import type {
  V1AlertSpec,
  V1ReportSpec,
} from "@rilldata/web-common/runtime-client";

/**
 * Returns the URL that opens a canvas report, with the canvas state captured at scheduling time applied.
 * Returns undefined if the report isn't for a canvas.
 */
export function getCanvasReportOpenUrl(
  reportSpec: V1ReportSpec,
  pathContext: DashboardPathContext,
): string | undefined {
  const canvasName = reportSpec.annotations?.["canvas"];
  if (!canvasName) return undefined;

  const path = getDashboardPath(pathContext, "canvas", canvasName);
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
