import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";

export async function load({ parent, url, params }) {
  const { report } = await parent();
  const organization = params.organization;
  const project = params.project;
  const reportId = params.report;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");
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
