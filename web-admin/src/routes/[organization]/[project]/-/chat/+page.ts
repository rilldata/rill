import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params: { organization, project } }) => {
  // NOTE: in the future, we'll use user-level `ai` permissions to determine access to the chat page.
  if (!get(featureFlags.chat)) {
    throw redirect(307, `/${organization}/${project}/-/dashboards`);
  }
};
