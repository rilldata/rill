import { runtimeServiceGetConfig } from "@rilldata/web-common/runtime-client/manual-clients";
import { error } from "@sveltejs/kit";

export const ssr = false;

/** @type {import('./$types').PageLoad} */
export async function load({ params, url }) {
  /** If ?focus, tell the page to focus the editor as soon as available */
  const focusEditor = url.searchParams.get("focus") === "";
  const config = await runtimeServiceGetConfig();
  if (config.readonly) {
    throw error(404, "Page not found");
  }

  return {
    modelName: params.name,
    focusEditor,
  };
}
