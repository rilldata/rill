import { error, redirect } from "@sveltejs/kit";
import { ExplorerMetricsDefinitionDoesntExist } from "@rilldata/web-local/common/errors/ErrorMessages";
import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");
  try {
    const dashboardMeta = await runtimeServiceGetFile(
      instanceResp.repoId,
      params.name
    );

    // TODO fix names
    const dashboardYAML = dashboardMeta.blob;

    // check if metrics definition is defined
    if (dashboardYAML.timeDimension !== undefined) {
      return {
        metricsDefId: params.name,
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
