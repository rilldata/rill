import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";

export async function load({ parent, url, params }) {
  const { alert } = await parent();
  const organization = params.organization;
  const project = params.project;
  const alertId = params.alert;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");
  const exploreName =
    alert.alert.spec.annotations["explore"] ??
    getExploreName(alert.alert.spec.annotations?.web_open_path); // backwards compatibility

  return {
    organization,
    project,
    alertId,
    alert,
    executionTime,
    token,
    exploreName,
  };
}
