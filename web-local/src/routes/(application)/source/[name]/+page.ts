import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params }) {
  const config = await runtimeServiceGetConfig();
  if (config.readonly) {
    throw error(404, "Page not found");
  }

  return { sourceName: params.name };
}
