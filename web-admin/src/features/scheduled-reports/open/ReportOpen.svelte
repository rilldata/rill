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
  import { toStore } from "svelte/store";
  import {
    type DashboardPathContext,
    getSpecQuery,
  } from "@rilldata/web-common/features/explore-mappers/utils";
  import type { V1ReportSpec } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    reportSpec,
    exploreName,
    executionTime,
    pathContext,
    ignoreFilters,
    backHref,
  }: {
    reportSpec: V1ReportSpec;
    exploreName: string;
    executionTime: string | null;
    pathContext: DashboardPathContext;
    // Set when the filters are already applied by the access token, to avoid applying them twice.
    ignoreFilters: boolean;
    backHref: string;
  } = $props();

  const runtimeClient = useRuntimeClient();

  const query = $derived(getSpecQuery(reportSpec));

  // Reports without a query can't be mapped to a dashboard, so open the report page instead.
  $effect(() => {
    if (!query.queryName || !query.queryArgsJson) {
      void goto(backHref, { replaceState: true });
    }
  });

  // A store of the options keeps the query in sync with the props, e.g. when the route is reused for another report.
  const exploreUrlQuery = createQuery(
    toStore(() =>
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
            exploreProtoState: reportSpec.annotations?.web_open_state,
            forceOpenPivot: true,
            ignoreFilters,
          },
        },
        pathContext,
      ),
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
        <h2 class="text-lg font-semibold">{m.report_unable_to_open()}</h2>
        <CtaMessage>
          {$exploreUrlQuery.error.message}
        </CtaMessage>
      </div>
      <CtaButton variant="secondary" href={backHref}>
        {m.report_go_to_page()}
      </CtaButton>
    {:else}
      <div class="mt-10">
        <LoadingSpinner />
      </div>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
