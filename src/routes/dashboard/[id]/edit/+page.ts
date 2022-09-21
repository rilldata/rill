import { getMetricsViewMetadata } from "$lib/svelte-query/queries/metrics-views/metadata";
import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  let metricsDefExists: boolean;
  await getMetricsViewMetadata(params.id).then((meta) => {
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

  throw error(404, "Metrics definition not found");
}
