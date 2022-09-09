import { error } from "@sveltejs/kit";

/** @type {import('./$types').PageLoad} */
export function load({ params }) {
  // TODO: Check to see if the metricsDefId exists server-side
  const metricsDefExists = true;

  if (metricsDefExists) {
    // TODO: should I return an object or does a string work?
    return {
      metricsDefId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Metrics definition not found");
}
