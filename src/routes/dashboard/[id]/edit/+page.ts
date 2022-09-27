import { config } from "$lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "$lib/svelte-query/queries/metrics-views/metadata";
import { error } from "@sveltejs/kit";
import {
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
} from "../../../../common/errors/ErrorMessages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const invalidDashboardErrors = [
    ExplorerSourceModelDoesntExist,
    ExplorerSourceModelIsInvalid,
    ExplorerTimeDimensionDoesntExist,
  ];

  try {
    const meta = await getMetricsViewMetadata(config, params.id);

    // check to see if metric definition exists
    if (meta.timeDimension !== undefined) {
      return {
        metricsDefId: params.id,
      };
    }
  } catch (err) {
    // the component will display invalid dashboard errors
    if (invalidDashboardErrors.includes(err.message)) {
      return {
        metricsDefId: params.id,
      };
    }
  }

  throw error(404, "Metrics definition not found");
}
