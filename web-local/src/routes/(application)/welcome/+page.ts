import { isProjectInitialized } from "@rilldata/web-common/features/welcome/is-project-initialized";
import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types.js";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";

export const load: PageLoad = async ({ parent }) => {
  const parentData = await parent();
  const initialized = await isProjectInitialized(
    queryClient,
    parentData.instanceId,
  );

  if (!initialized) return;
  throw redirect(303, "/");
};
