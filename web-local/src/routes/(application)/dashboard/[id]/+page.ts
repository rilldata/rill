import { config } from "@rilldata/web-local/lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "@rilldata/web-local/lib/svelte-query/queries/metrics-views/metadata";
import { error, redirect } from "@sveltejs/kit";
import { ExplorerDashboardDoesntExist } from "@rilldata/web-local/common/errors/ErrorMessages";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  try {
    const meta = await getMetricsViewMetadata(config, params.id);

    // check if metrics definition is defined
    if (meta.timeDimension !== undefined) {
      return {
        metricsDefId: params.id,
      };
    }

    // if metrics definition is not yet defined, redirect to the metrics definition page
    if (meta.timeDimension === undefined) {
      return redirect(307, `/dashboard/${params.id}/edit`);
    }
  } catch (err) {
    if (
      ExplorerDashboardDoesntExist.includes(err.message) ||
      err.message.includes(ExplorerDashboardDoesntExist)
    ) {
      throw error(404, "Dashboard not found");
    } else {
      throw redirect(307, `/dashboard/${params.id}/edit`);
    }
  }
  throw redirect(307, `/dashboard/${params.id}/edit`);
}
