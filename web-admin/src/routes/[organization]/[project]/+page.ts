import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({ params: { organization, project }, parent }) => {
  // Wait for the feature flags to load
  await parent();

  // NOTE: in the future, we'll use user-level `ai` permissions to determine access to the chat page.
  const chatEnabled = get(featureFlags.chat);
  if (chatEnabled) {
    throw redirect(307, `/${organization}/${project}/-/chat`);
  }

  throw redirect(307, `/${organization}/${project}/-/dashboards`);
};
