export async function load({ parent, params }) {
  const { report } = await parent();
  const organization = params.organization;
  const project = params.project;
  const reportName = params.report;

  return {
    organization,
    project,
    reportName,
    report,
  };
}
