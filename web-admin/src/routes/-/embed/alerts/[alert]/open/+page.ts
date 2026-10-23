import { fetchEmbedResource } from "@rilldata/web-admin/features/embeds/fetch-embed-resource";
import { getOpenExploreName } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

export async function load({ parent, url, params }) {
  const alert = await fetchEmbedResource(
    await parent(),
    ResourceKind.Alert,
    params.alert,
  );
  const alertSpec = alert?.alert?.spec ?? {};

  return {
    alertId: params.alert,
    alertSpec,
    executionTime: url.searchParams.get("execution_time"),
    exploreName: getOpenExploreName(alertSpec),
  };
}
