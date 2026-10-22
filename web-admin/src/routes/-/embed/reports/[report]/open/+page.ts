import { fetchEmbedResource } from "@rilldata/web-admin/features/embeds/fetch-embed-resource";
import {
  EmbedDashboardPaths,
  getCanvasReportOpenUrl,
  getOpenExploreName,
} from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { isAIReportSpec } from "@rilldata/web-common/features/scheduled-reports/utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const report = await fetchEmbedResource(
    await parent(),
    ResourceKind.Report,
    params.report,
  );
  const reportSpec = report?.report?.spec ?? {};

  // Canvas reports open the canvas dashboard directly, with the canvas state captured at scheduling time applied.
  const canvasUrl = getCanvasReportOpenUrl(reportSpec, EmbedDashboardPaths);
  if (canvasUrl) {
    redirect(307, canvasUrl);
  }

  return {
    reportId: params.report,
    reportSpec,
    // AI reports open an AI conversation, which embeds don't support.
    isAIReport: isAIReportSpec(reportSpec),
    executionTime: url.searchParams.get("execution_time"),
    exploreName: getOpenExploreName(reportSpec),
  };
}
