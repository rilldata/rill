import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import { stripInternalReportParams } from "@rilldata/web-common/features/scheduled-reports/utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { report } = await parent();
  const organization = params.organization;
  const project = params.project;
  const reportId = params.report;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");

  // Canvas reports open the canvas dashboard directly, with the canvas state
  // (URL search params) captured at scheduling time applied.
  const canvasName = report.report.spec.annotations["canvas"];
  if (canvasName) {
    const path = token
      ? `/${organization}/${project}/-/share/${token}/canvas/${canvasName}`
      : `/${organization}/${project}/canvas/${canvasName}`;
    const stateParams = stripInternalReportParams(
      new URLSearchParams(report.report.spec.annotations?.web_open_state ?? ""),
    );
    const search = stateParams.toString();
    redirect(307, search ? `${path}?${search}` : path);
  }

  const exploreName =
    report.report.spec.annotations["explore"] ??
    getExploreName(report.report.spec.annotations?.web_open_path); // backwards compatibility

  return {
    organization,
    project,
    reportId,
    report,
    executionTime,
    token,
    exploreName,
  };
}
