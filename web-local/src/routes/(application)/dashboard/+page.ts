import { error, redirect } from "@sveltejs/kit";
import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { get } from "svelte/store";

export function load() {
  const readyOnly = get(featureFlags.readOnly);

  if (readyOnly) {
    throw error(404, `File not found`);
  }
  throw redirect(307, `/`);
}
