export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  isProjectInitialized,
  handleUninitializedProject,
} from "@rilldata/web-common/features/welcome/is-project-initialized";
import { get } from "svelte/store";

export async function load({ url, depends, untrack }) {
  depends("init");

  const instanceId = get(runtime).instanceId;
  const initialized = await isProjectInitialized(instanceId);

  const onWelcomePage = untrack(() => url.pathname === "/welcome");

  if (!initialized) {
    const goToWelcomePage = await handleUninitializedProject(instanceId);
    if (goToWelcomePage && !onWelcomePage) throw redirect(303, "/welcome");
  } else if (onWelcomePage) {
    throw redirect(303, "/");
  }
}
