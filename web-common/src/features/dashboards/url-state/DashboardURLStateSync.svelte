<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { getUrlFromMetricsExplorer } from "@rilldata/web-common/features/dashboards/url-state/toUrl";
  import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let partialMetrics: Partial<MetricsExplorerEntity>;
  export let defaultPartialMetrics: Partial<MetricsExplorerEntity> | undefined =
    undefined;

  const {
    metricsViewName,
    exploreName,
    dashboardStore,
    validSpecStore,
    timeRangeSummaryStore,
  } = getStateManagers();
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    $runtime.instanceId,
    $metricsViewName,
  );
  $: ({ error: schemaError } = $metricsViewSchema);

  $: ({ error, data: timeRangeSummary } = $timeRangeSummaryStore);
  $: timeRangeSummaryError = error as HTTPError;

  let prevUrl = "";
  function gotoNewState() {
    if (!exploreSpec) return;

    const u = new URL(
      `${$page.url.protocol}//${$page.url.host}${$page.url.pathname}`,
    );
    getUrlFromMetricsExplorer(
      $dashboardStore,
      u.searchParams,
      exploreSpec,
      exploreSpec.defaultPreset ?? {},
    );
    const newUrl = u.toString();
    if (window.location.href !== newUrl) {
      void goto(newUrl);
    }
  }

  function mergePartialExplorerEntity() {
    if (!metricsSpec || !exploreSpec) return;
    if (!$dashboardStore) {
      // initial page load, create an entry in metricsExplorerStore
      metricsExplorerStore.init(
        $exploreName,
        metricsSpec,
        exploreSpec,
        timeRangeSummary,
      );
      // if a default metrics is passed sync with it then merge with it.
      // this could be a bookmark or a magic auth token state.
      if (defaultPartialMetrics) {
        metricsExplorerStore.mergePartialExplorerEntity(
          $exploreName,
          defaultPartialMetrics,
          metricsSpec,
        );
      }
    }

    console.log(partialMetrics);
    metricsExplorerStore.mergePartialExplorerEntity(
      $exploreName,
      partialMetrics,
      metricsSpec,
    );
    prevUrl = window.location.href;
  }

  // reactive to only dashboardStore
  // but gotoNewState checks other fields
  $: if ($dashboardStore) {
    gotoNewState();
  }

  // reactive to only partialMetrics, metricsSpec & exploreSpec
  // but mergePartialExplorerEntity checks other fields
  $: if (
    partialMetrics &&
    metricsSpec &&
    exploreSpec &&
    prevUrl !== window.location.href
  ) {
    mergePartialExplorerEntity();
  }
</script>

{#if schemaError}
  <ErrorPage
    statusCode={schemaError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the schema for this dashboard."
    detail={schemaError?.response?.data?.message}
  />
{:else if timeRangeSummaryError}
  <ErrorPage
    statusCode={timeRangeSummaryError?.response?.status}
    header="Error loading dashboard"
    body="Unable to fetch the time range for this dashboard."
    detail={timeRangeSummaryError?.response?.data?.message}
  />
{:else if $dashboardStore}
  <slot />
{/if}
