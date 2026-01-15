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
import { localServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
import { Settings } from "luxon";

Settings.defaultLocale = "en";

export async function load({ url, depends, untrack }) {
  depends("app:init");

  // Fetch metadata to check preview mode
  const metadata = await localServiceGetMetadata();
  const previewMode = metadata.previewMode ?? false;

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
    if (!url.searchParams.get("redirect")) return false;

    // In preview mode, redirect to /preview instead of /files
    if (previewMode) {
      return url.pathname !== "/home" && "/home";
    }

    return (
      url.pathname !== `/files${firstDashboardFile?.path}` &&
      `/files${firstDashboardFile?.path}`
    );
  });

  if (!initialized) {
    initialized = await handleUninitializedProject(instanceId);
  } else {
    // Only redirect after project is initialized
    if (previewMode && url.pathname === "/") {
      throw redirect(303, "/home");
    }
    if (redirectPath) {
      throw redirect(303, redirectPath);
    }
  }

  return { initialized, previewMode };
}
