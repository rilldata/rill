import { fetchResource } from "@rilldata/web-admin/features/resources/fetch-resource";
import { getOpenExploreName } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2/context";
import { error } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  // Embed routes don't have the project layout, so the runtime client is built from the embed layout's data.
  // The embed layout shows an error for missing params, so there's nothing to fetch.
  const { instanceId, runtimeHost, accessToken } = await parent();
  if (!instanceId || !runtimeHost || !accessToken) throw error(400);
  const client = getRuntimeClient({
    host: runtimeHost,
    instanceId,
    jwt: accessToken,
    authContext: "embed",
  });
  const alert = await fetchResource(client, ResourceKind.Alert, params.alert);
  const alertSpec = alert?.alert?.spec ?? {};

  return {
    alertId: params.alert,
    alertSpec,
    executionTime: url.searchParams.get("execution_time"),
    exploreName: getOpenExploreName(alertSpec),
  };
}
