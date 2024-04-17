import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";
import { redirect } from "@sveltejs/kit";

export async function load({ parent }) {
  const parentData = await parent();
  const initialized = await isProjectInitialized(parentData.instanceId);

  if (!initialized) return;
  throw redirect(303, "/");
}
