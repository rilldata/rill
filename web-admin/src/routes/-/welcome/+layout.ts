import { UserWelcomeStatus } from "@rilldata/web-admin/features/welcome/welcom-store.ts";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async () => {
  if (!get(UserWelcomeStatus)) {
    throw redirect(307, "/");
  }
};
