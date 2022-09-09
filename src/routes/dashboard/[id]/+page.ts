import { error } from "@sveltejs/kit";

// TODO: figure out what this type does
/** @type {import('./$types').PageLoad} */
export function load({ params }) {
  // TOOD: Check to see if the metricsDefId exists server-side
  const metricsDefExists = true;

  if (metricsDefExists) {
    // TODO: should I return an object or does a string work?
    return {
      metricsDefId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Dashboard not found");
}
