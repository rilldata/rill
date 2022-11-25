import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import {
  ExplorerSourceColumnDoesntExist,
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
  ExplorerMetricsDefinitionDoesntExist,
} from "@rilldata/web-local/common/errors/ErrorMessages";
import { fetchWrapper } from "@rilldata/web-local/lib/util/fetchWrapper";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const instanceResp = await fetchWrapper("v1/runtime/instance-id", "GET");
  try {
    const dashboardMeta = await runtimeServiceGetFile(
      instanceResp.instanceId,
      `dashboards/${params.name}.yaml`
    );

    const dashboardYAML = dashboardMeta.blob;

    // if metric definition exists, go to component
    if (dashboardYAML) {
      return {
        metricsDefName: params.name,
        yaml: dashboardYAML,
      };
    }
  } catch (err) {
    const invalidDashboardErrors = [
      ExplorerSourceModelDoesntExist,
      ExplorerSourceModelIsInvalid,
      ExplorerSourceColumnDoesntExist,
      ExplorerTimeDimensionDoesntExist,
    ];

    // any invalid dashboard error will be displayed by the component
    if (
      invalidDashboardErrors.some(
        (errMsg) => errMsg.includes(err.message) || err.message.includes(errMsg)
      )
    ) {
      return {
        metricsDefId: params.name,
      };
    } else {
      if (
        ExplorerMetricsDefinitionDoesntExist.includes(err.message) ||
        err.message.includes(ExplorerMetricsDefinitionDoesntExist)
      ) {
        throw error(404, "Metrics definition  not found");
      }
      // Pass non standard error message to be shown in dialog
      return {
        metricsDefId: params.name,
        error: err.message,
      };
    }
  }

  throw error(404, "Metrics definition not found");
}
