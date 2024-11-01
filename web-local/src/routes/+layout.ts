export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client/index.js";
import { handleUninitializedProject } from "@rilldata/web-common/features/welcome/is-project-initialized.js";

export async function load({ url, depends, untrack }) {
  depends("init");

  const instanceId = get(runtime).instanceId;

  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(instanceId, undefined, signal);
    },
  });

  const firstDashboardFile = files.files?.find((file) =>
    file.path?.startsWith("/dashboards/"),
  );

  let initialized = !!files.files?.some(({ path }) => path === "/rill.yaml");

  const redirectPath = untrack(() => {
    return (
      !!url.searchParams.get("redirect") &&
      url.pathname !== `/files${firstDashboardFile?.path}` &&
      `/files${firstDashboardFile?.path}`
    );
  });

  if (!initialized) {
    initialized = await handleUninitializedProject(instanceId);
  } else if (redirectPath) {
    throw redirect(303, redirectPath);
  }

  return { initialized };
}
