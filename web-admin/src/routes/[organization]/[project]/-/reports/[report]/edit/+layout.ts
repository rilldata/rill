import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils.ts";

export async function load({ parent, params }) {
  const { report } = await parent();
  const organization = params.organization;
  const project = params.project;
  const reportName = params.report;
  const reportSpec = report?.report?.spec;

  const exploreName =
    reportSpec.annotations?.["explore"] ??
    getExploreName(reportSpec.annotations?.web_open_path ?? "");

  return {
    organization,
    project,
    reportName,
    report,
    exploreName,
  };
}
