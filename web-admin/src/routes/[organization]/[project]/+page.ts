import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params: { organization, project }, parent }) => {
  if (get(featureFlags.chat)) {
    throw redirect(307, `/${organization}/${project}/-/chat`);
  }

  // TODO: if the user has `ai` permissions, redirect to `/${organization}/${project}/-/chat`
  // otherwise, redirect to `/${organization}/${project}/-/dashboards`
  throw redirect(307, `/${organization}/${project}/-/dashboards`);
};
