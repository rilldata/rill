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

export async function load({ url }) {
  // depends("init"); // Removing for now, but reconsider later.
  const instanceId = get(runtime).instanceId;

  // Redirect to the welcome page if the project is not initialized
  const onboardingState = getOnboardingState(); // TODO: Make sure this doesn't trigger an unnecessary fetch of `onboarding-state.json`
  const initialized = await onboardingState.isInitialized();
  const inOnboardingFlow = url.pathname.startsWith("/welcome");

  if (!initialized && !inOnboardingFlow) {
    throw redirect(303, "/welcome");
  }

  // If the user has clicked on an example project, redirect to the project's first dashboard
  const shouldRedirectToFirstDashboard = !!url.searchParams.get("redirect");

  if (shouldRedirectToFirstDashboard) {
    const files = await queryClient.fetchQuery<V1ListFilesResponse>({
      queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
      queryFn: ({ signal }) => {
        return runtimeServiceListFiles(instanceId, undefined, signal);
      },
    });

    const firstDashboardFile = files.files?.find((file) =>
      file.path?.startsWith("/dashboards/"),
    );

    if (url.pathname !== `/files${firstDashboardFile?.path}`) {
      throw redirect(303, `/files${firstDashboardFile?.path}`);
    }
  }

  return { onboardingState };
}
