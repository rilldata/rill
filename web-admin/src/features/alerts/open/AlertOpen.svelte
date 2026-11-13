<script lang="ts">
  import { goto } from "$app/navigation";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/explore-mappers/map-to-explore";
  import {
    type DashboardPathContext,
    getDashboardPath,
    getExplorePageUrlSearchParams,
    getSpecQuery,
  } from "@rilldata/web-common/features/explore-mappers/utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { V1AlertSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    alertSpec,
    exploreName,
    executionTime,
    pathContext,
    ignoreFilters,
    backHref,
  }: {
    alertSpec: V1AlertSpec;
    exploreName: string;
    executionTime: string | null;
    pathContext: DashboardPathContext;
    // Set when the filters are already applied by the access token, to avoid applying them twice.
    // Applying threshold filters twice leads to incorrect results.
    ignoreFilters: boolean;
    backHref: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let query = $derived(getSpecQuery(alertSpec));

  let dashboardState = $derived(
    mapQueryToDashboard(
      runtimeClient,
      { exploreName, ...query, executionTime: executionTime ?? undefined },
      {
        exploreProtoState: alertSpec.annotations?.web_open_state,
        ignoreFilters,
      },
    ),
  );

  // Alerts without a query can't be mapped to a dashboard, so open the alert page instead.
  $effect(() => {
    if (!query.queryName || !query.queryArgsJson) {
      void goto(backHref, { replaceState: true });
    }
  });

  $effect(() => {
    const data = $dashboardState?.data;
    if (!data) return;
    void getExplorePageUrlSearchParams(
      runtimeClient,
      data.exploreName,
      data.exploreState,
    ).then((params) => {
      const path = getDashboardPath(pathContext, "explore", data.exploreName);
      const search = params.toString();
      // Replace this page in the history, so going back from the dashboard doesn't land here and redirect again.
      return goto(search ? `${path}?${search}` : path, { replaceState: true });
    });
  });
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $dashboardState.isLoading}
      <div class="mt-10">
        <LoadingSpinner />
      </div>
    {:else if $dashboardState.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.alert_unable_to_open()}</h2>
        <CtaMessage>
          {$dashboardState.error}
        </CtaMessage>
      </div>
      <CtaButton variant="secondary" href={backHref}>
        {m.alert_go_to_page()}
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
