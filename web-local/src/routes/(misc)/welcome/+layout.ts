import { WelcomeStatus } from "@rilldata/web-common/features/welcome/status.ts";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export async function load({ parent }) {
  const { initialized, previewMode } = await parent();
  // Project can get initialized during welcome steps.
  // We should only block this route by redirecting to home if all welcome steps are complete.
  if (initialized && !get(WelcomeStatus)) {
    throw redirect(303, previewMode ? "/dashboards" : "/");
  }
}
