import { fetchResource } from "@rilldata/web-admin/features/resources/fetch-resource";
import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.js";

export async function load({ params, parent }) {
  const { runtime } = await parent();
  return {
    alert: await fetchResource(
      getCloudRuntimeClient(runtime),
      ResourceKind.Alert,
      params.alert,
    ),
  };
}
