<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import { convertExploreStateToURLSearchParams } from "@rilldata/web-common/features/dashboards/url-state/convertExploreStateToURLSearchParams";
  import { getUpdatedUrlForExploreState } from "@rilldata/web-common/features/dashboards/url-state/getUpdatedUrlForExploreState";
  import {
    createQueryServiceMetricsViewSchema,
    type V1ExplorePreset,
  } from "@rilldata/web-common/runtime-client";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricsViewName: string;
  export let exploreName: string;
  export let defaultExplorePreset: V1ExplorePreset;
  export let initExploreState: Partial<MetricsExplorerEntity>;
  export let initUrlSearch: string;
  export let partialExploreState: Partial<MetricsExplorerEntity>;
  export let urlSearchForPartial: string;

  const { dashboardStore, validSpecStore, timeRangeSummaryStore } =
    getStateManagers();
  $: exploreSpec = $validSpecStore.data?.explore;
  $: metricsSpec = $validSpecStore.data?.metricsView;

  const metricsViewSchema = createQueryServiceMetricsViewSchema(
    $runtime.instanceId,
    metricsViewName,
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
    u.search = convertExploreStateToURLSearchParams(
      $dashboardStore,
      exploreSpec,
      defaultExplorePreset,
    );
    const newUrl = u.toString();
    if (window.location.href !== newUrl) {
      void goto(newUrl);
    }
  }

  function mergePartialExplorerEntity() {
    if (!metricsSpec || !exploreSpec) return;
    const isInit = !$dashboardStore;
    if (isInit) {
      // initial page load, create an entry in metricsExplorerStore
      metricsExplorerStore.init(
        exploreName,
        metricsSpec,
        exploreSpec,
        timeRangeSummary,
        initExploreState,
      );
    }

    if (
      // if not dashboard init then always merge partial
      !isInit ||
      // else during init only merge if partial updates the url params
      urlSearchForPartial !== ""
    ) {
      metricsExplorerStore.mergePartialExplorerEntity(
        exploreName,
        partialExploreState,
        metricsSpec,
      );
    }
    const curSearch = $page.url.searchParams.toString();
    const partialDifferentThanCurrent = curSearch !== urlSearchForPartial;
    const initDifferentThanCurrent = curSearch !== initUrlSearch;
    if (partialDifferentThanCurrent || (isInit && initDifferentThanCurrent)) {
      const redirectUrl = new URL($page.url);
      redirectUrl.search = partialDifferentThanCurrent
        ? urlSearchForPartial
        : initUrlSearch;
      history.replaceState(history.state, "", redirectUrl);
      prevUrl = redirectUrl.toString();
    } else {
      prevUrl = window.location.href;
    }
  }

  // reactive to only dashboardStore
  // but gotoNewState checks other fields
  $: if ($dashboardStore) {
    gotoNewState();
  }

  // reactive to only partialExploreState, metricsSpec & exploreSpec
  // but mergePartialExplorerEntity checks other fields
  $: if (
    partialExploreState &&
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
