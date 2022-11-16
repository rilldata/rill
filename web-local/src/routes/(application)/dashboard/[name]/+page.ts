import { error, redirect } from "@sveltejs/kit";
import { ExplorerMetricsDefinitionDoesntExist } from "@rilldata/web-local/common/errors/ErrorMessages";
import { runtimeServiceGetCatalogObject } from "@rilldata/web-common/runtime-client";
import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");
  try {
    const dashboardResp = await runtimeServiceGetCatalogObject(
      instanceResp.instanceId,
      params.name
    );

    const dashboardMeta = dashboardResp.object.metricsView;

    console.log(dashboardMeta);

    // check if metrics definition is defined
    if (dashboardMeta.timeDimension !== undefined) {
      return {
        metricsDefId: dashboardMeta.name,
      };
    }

    // if metrics definition is not yet defined, redirect to the metrics definition page
    if (dashboardMeta.timeDimension === undefined) {
      return redirect(307, `/dashboard/${params.name}/edit`);
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
