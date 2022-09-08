import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
import { error } from "@sveltejs/kit";

// TODO: figure out what this type does
/** @type {import('./$types').PageLoad} */
export function load({ params }) {
  // Check to see if the metricsDefId exists server-side
  const metricsDefinition = getMetricsDefReadableById(params.id);

  let metricsDefExists: boolean;
  metricsDefinition.subscribe((metricsDef) => {
    // Q: Why is this undefined?
    console.log("metricsDef", metricsDef);
    metricsDefExists = !!metricsDef?.id;
  });

  if (true) {
    // TODO: should I return an object or does a string work?
    return {
      metricsDefId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Dashboard not found");
}
