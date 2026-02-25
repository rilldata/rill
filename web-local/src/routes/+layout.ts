export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { V1ListFilesResponse } from "@rilldata/web-common/runtime-client/index.js";
import { handleUninitializedProject } from "@rilldata/web-common/features/welcome/is-project-initialized.js";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { LOCAL_HOST, LOCAL_INSTANCE_ID } from "../lib/local-runtime-config";
import { Settings } from "luxon";

Settings.defaultLocale = "en";

export async function load({ url, depends, untrack }) {
  depends("init");

  const instanceId = LOCAL_INSTANCE_ID;

  const client = new RuntimeClient({
    host: LOCAL_HOST,
    instanceId,
  });

  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(client, {}, { signal });
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
    initialized = await handleUninitializedProject(client);
  } else if (redirectPath) {
    throw redirect(303, redirectPath);
  }

  return { initialized };
}
