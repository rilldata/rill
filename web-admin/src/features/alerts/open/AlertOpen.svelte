<script lang="ts">
  import { goto } from "$app/navigation";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import { getMappedExploreUrlQueryOptions } from "@rilldata/web-common/features/explore-mappers/url-mappers/get-mapped-url.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { createQuery } from "@tanstack/svelte-query";
  import {
    type DashboardPathContext,
    getSpecQuery,
  } from "@rilldata/web-common/features/explore-mappers/utils";
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

  // svelte-ignore state_referenced_locally
  const query = getSpecQuery(alertSpec);

  // Alerts without a query can't be mapped to a dashboard, so open the alert page instead.
  $effect(() => {
    if (!query.queryName || !query.queryArgsJson) {
      void goto(backHref, { replaceState: true });
    }
  });

  // svelte-ignore state_referenced_locally
  const exploreUrlQuery = createQuery(
    getMappedExploreUrlQueryOptions(
      runtimeClient,
      {
        type: "query",
        request: {
          exploreName,
          ...query,
          executionTime: executionTime ?? undefined,
        },
        options: {
          exploreProtoState: alertSpec.annotations?.web_open_state,
          ignoreFilters,
        },
      },
      pathContext,
    ),
  );

  $effect(() => {
    const url = $exploreUrlQuery.data;
    if (!url) return;
    // Replace this page in the history, so going back from the dashboard doesn't land here and redirect again.
    void goto(url, { replaceState: true });
  });
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $exploreUrlQuery.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.alert_unable_to_open()}</h2>
        <CtaMessage>
          {$exploreUrlQuery.error.message}
        </CtaMessage>
      </div>
      <CtaButton variant="secondary" href={backHref}>
        {m.alert_go_to_page()}
      </CtaButton>
    {:else}
      <div class="mt-10">
        <LoadingSpinner />
      </div>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
