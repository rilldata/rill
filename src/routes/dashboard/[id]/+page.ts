import { config } from "$lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "$lib/svelte-query/queries/metrics-views/metadata";
import { error, redirect } from "@sveltejs/kit";
import {
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
  ExplorerTimeDimensionDoesntExist,
} from "../../../common/errors/ErrorMessages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const invalidDashboardErrors = [
    ExplorerSourceModelDoesntExist,
    ExplorerSourceModelIsInvalid,
    ExplorerTimeDimensionDoesntExist,
  ];

  try {
    const meta = await getMetricsViewMetadata(config, params.id);

    // check to see if metrics definition exists
    if (meta.timeDimension !== undefined) {
      return {
        metricsDefId: params.id,
      };
    }
  } catch (err) {
    // check to see if dashboard is valid
    if (invalidDashboardErrors.includes(err.message)) {
      throw redirect(307, `/dashboard/${params.id}/edit`);
    }
  }

  throw error(404, "Dashboard not found");
}
