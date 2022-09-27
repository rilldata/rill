import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "@rilldata/web-local/lib/svelte-query/queries/metrics-views/metadata";
import { error, redirect } from "@sveltejs/kit";
import {
  ExplorerSourceColumnDoesntExist,
  ExplorerSourceModelDoesntExist,
  ExplorerSourceModelIsInvalid,
} from "../../../common/errors/ErrorMessages";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const meta = await getMetricsViewMetadata(config, params.id);

    // check to see if metrics definition exists
    if (meta.timeDimension !== undefined) {
      return {
        metricsDefId: params.id,
      };
    }
  } catch (err) {
    const invalidDashboardErrors = [
      ExplorerSourceModelDoesntExist,
      ExplorerSourceModelIsInvalid,
      ExplorerSourceColumnDoesntExist,
    ];

    // if dashboard is invalid, redirect to the metrics definition page
    if (
      invalidDashboardErrors.some(
        (errMsg) => errMsg.includes(err.message) || err.message.includes(errMsg)
      )
    ) {
      throw redirect(307, `/dashboard/${params.id}/edit`);
    }
  }

  throw error(404, "Dashboard not found");
}
