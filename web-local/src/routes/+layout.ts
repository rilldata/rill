export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  isProjectInitialized,
  handleUninitializedProject,
  firstLoad,
} from "@rilldata/web-common/features/welcome/is-project-initialized";
import { get } from "svelte/store";

export async function load({ url: { pathname }, depends }) {
  depends("init");

  // Remove this and untrack URL changes
  // after upgrading to SvelteKit 2.0
  if (!get(firstLoad)) return;
  firstLoad.set(false);

  const instanceId = get(runtime).instanceId;
  const initialized = await isProjectInitialized(instanceId);

  const onWelcomePage = pathname === "/welcome";

  if (!initialized) {
    const goToWelcomePage = await handleUninitializedProject(instanceId);
    if (goToWelcomePage && !onWelcomePage) throw redirect(303, "/welcome");
  } else if (onWelcomePage) {
    throw redirect(303, "/");
  }
}
