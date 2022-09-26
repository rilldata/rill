import { config } from "$web-local/lib/application-state-stores/application-store";
import { getMetricsViewMetadata } from "$web-local/lib/svelte-query/queries/metrics-views/metadata";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  let metricsDefExists: boolean;

  await getMetricsViewMetadata(config, params.id).then((meta) => {
    if (meta.timeDimension !== undefined) {
      metricsDefExists = true;
    } else {
      metricsDefExists = false;
    }
  });

  if (metricsDefExists) {
    return {
      metricsDefId: params.id,
    };
  }

  // TODO: determine when the dashboard is invalid, then redirect
  // if (dashboardInvalid) {
  //   throw redirect(307, `/dashboard/${params.id}/edit`);
  // }

  throw error(404, "Dashboard not found");
}
