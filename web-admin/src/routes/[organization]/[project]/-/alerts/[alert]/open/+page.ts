import {
  isResourceErrorAlert,
  RESOURCE_ERRORS_OPEN_PATH,
} from "@rilldata/web-admin/features/alerts/resource-errors/utils";
import { getOpenExploreName } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { alert } = await parent();
  const alertSpec = alert?.alert?.spec ?? {};
  if (isResourceErrorAlert(alertSpec)) {
    throw redirect(
      307,
      `/${params.organization}/${params.project}${RESOURCE_ERRORS_OPEN_PATH}`,
    );
  }
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
