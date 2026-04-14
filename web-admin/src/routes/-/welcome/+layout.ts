import { InWelcomeFlowStore } from "@rilldata/web-admin/features/welcome/welcome-store.ts";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async () => {
  if (!get(InWelcomeFlowStore)) {
    throw redirect(307, "/");
  }
};
