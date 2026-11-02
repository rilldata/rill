import {
  getCanvasReportOpenUrl,
  getCloudDashboardPaths,
  getOpenExploreName,
} from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { isAIReportSpec } from "@rilldata/web-common/features/scheduled-reports/utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { report } = await parent();
  const organization = params.organization;
  const project = params.project;
  const reportId = params.report;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");
  const reportSpec = report?.report?.spec ?? {};

  // Canvas reports open the canvas dashboard directly, with the canvas state
  // (URL search params) captured at scheduling time applied.
  const canvasUrl = getCanvasReportOpenUrl(
    reportSpec,
    getCloudDashboardPaths(organization, project, token),
  );
  if (canvasUrl) {
    redirect(307, canvasUrl);
  }

  // AI reports open the AI conversation created by the report run.
  // The runtime adds the session ID to the open link. In creator mode the link also carries a magic token,
  // which grants read access to the shared conversation and must be forwarded so the conversation page can authenticate with it.
  if (isAIReportSpec(reportSpec)) {
    const sessionId = url.searchParams.get("session_id");
    const path = sessionId
      ? `/${organization}/${project}/-/ai/${sessionId}`
      : `/${organization}/${project}/-/ai`;
    const search = token ? new URLSearchParams({ token }).toString() : "";
    redirect(307, search ? `${path}?${search}` : path);
  }

  return {
    organization,
    project,
    reportId,
    reportSpec,
    executionTime,
    token,
    exploreName: getOpenExploreName(reportSpec),
  };
}
