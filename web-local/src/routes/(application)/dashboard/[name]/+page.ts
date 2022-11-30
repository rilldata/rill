import { runtimeServiceGetCatalogEntry } from "@rilldata/web-common/runtime-client";
import { ExplorerMetricsDefinitionDoesntExist } from "@rilldata/web-local/common/errors/ErrorMessages";
import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
import { error, redirect } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");
  try {
    const dashboardMeta = await runtimeServiceGetCatalogEntry(
      instanceResp.instanceId,
      params.name
    );

    const dashboardYAML = dashboardMeta?.entry?.metricsView;

    // check if metrics definition is defined
    if (dashboardYAML) {
      return {
        metricViewName: params.name,
      };
    }
  } catch (err) {
    if (
      ExplorerMetricsDefinitionDoesntExist.includes(err.message) ||
      err.message.includes(ExplorerMetricsDefinitionDoesntExist)
    ) {
      throw error(404, "Dashboard not found");
    } else {
      throw redirect(307, `/dashboard/${params.name}/edit`);
    }
  }
  throw redirect(307, `/dashboard/${params.name}/edit`);
}
