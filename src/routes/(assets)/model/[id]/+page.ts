import { error } from "@sveltejs/kit";

// TODO: figure out what this type does
/** @type {import('./$types').PageLoad} */
export function load({ params }) {
  // TODO: Check to see if the modelId exists server-side
  const modelExists = true;

  if (modelExists) {
    // TODO: should I return an object or does a string work?
    return {
      modelId: params.id,
    };
  }

  console.log("params", params);
  throw error(404, "Model not found");
}
