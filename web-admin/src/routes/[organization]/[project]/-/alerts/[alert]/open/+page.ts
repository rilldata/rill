import { getOpenExploreName } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";

export async function load({ parent, url, params }) {
  const { alert } = await parent();
  const alertSpec = alert?.alert?.spec ?? {};

  return {
    organization: params.organization,
    project: params.project,
    alertId: params.alert,
    alertSpec,
    executionTime: url.searchParams.get("execution_time"),
    token: url.searchParams.get("token"),
    exploreName: getOpenExploreName(alertSpec),
  };
}
