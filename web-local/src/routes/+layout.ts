export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { projectInitialized } from "@rilldata/web-common/runtime-client/runtime-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";
import {
  isProjectInitialized,
  handleUninitializedProject,
} from "@rilldata/web-common/features/welcome/is-project-initialized.js";
import { get } from "svelte/store";

let firstLoad = true;

export async function load({ url: { pathname } }) {
  // Remove this and untrack URL changes
  // after upgrading to SvelteKit 2.0
  if (!firstLoad) return;
  firstLoad = false;

  const instanceId = get(runtime).instanceId;
  const initialized = await isProjectInitialized(instanceId);

  projectInitialized.set(!!initialized);

  const onWelcomePage = pathname === "/welcome";

  if (!initialized) {
    const goToWelcomePage = await handleUninitializedProject(instanceId);
    if (goToWelcomePage && !onWelcomePage) throw redirect(303, "/welcome");
  } else if (onWelcomePage) {
    throw redirect(303, "/");
  }
}
