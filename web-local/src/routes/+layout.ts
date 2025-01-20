export const ssr = false;

import { getOnboardingState } from "@rilldata/web-common/features/welcome/wizard/onboarding-state";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client/index.js";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

// TODO: Move initilization logic to the OnboardingState class
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

  const initialized = getOnboardingState().isInitialized();

  const redirectPath = untrack(() => {
    return (
      !!url.searchParams.get("redirect") &&
      url.pathname !== `/files${firstDashboardFile?.path}` &&
      `/files${firstDashboardFile?.path}`
    );
  });

  if (!initialized) {
    // initialized = await handleUninitializedProject(instanceId);
  } else if (redirectPath) {
    throw redirect(303, redirectPath);
  }

  return { initialized };
}
