import { runtimeServiceGetFile } from "@rilldata/web-common/runtime-client";
import {
  ExplorerMetricsDefinitionDoesntExist,
  ExplorerSourceColumnDoesntExist,
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
} from "@rilldata/web-local/lib/temp/errors/ErrorMessages";
import { fetchWrapperDirect } from "@rilldata/web-local/lib/util/fetchWrapper";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const localConfig = await fetchWrapperDirect(
    `${RILL_RUNTIME_URL}/local/config`,
    "GET"
  );

  try {
    const dashboardMeta = await runtimeServiceGetFile(
      localConfig.instance_id,
      `dashboards/${params.name}.yaml`
    );

    const dashboardYAML = dashboardMeta.blob;

    // if metric definition exists, go to component
    if (dashboardYAML) {
      return {
        metricsDefName: params.name,
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
        metricsName: params.name,
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
        metricsName: params.name,
        error: err.message,
      };
    }
  }

  throw error(404, "Metrics definition not found");
}
