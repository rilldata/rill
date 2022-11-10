import {
  ExplorerSourceColumnDoesntExist,
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
  ExplorerMetricsDefinitionDoesntExist,
} from "@rilldata/web-local/common/errors/ErrorMessages";
import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "@rilldata/web-local/lib/svelte-query/queries/metrics-views/metadata";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const meta = await getMetricsViewMetadata(config, params.id);

    // if metric definition exists, go to component
    if (meta) {
      return {
        metricsDefId: params.id,
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
        metricsDefId: params.id,
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
        metricsDefId: params.id,
        error: err.message,
      };
    }
  }

  throw error(404, "Metrics definition not found");
}
