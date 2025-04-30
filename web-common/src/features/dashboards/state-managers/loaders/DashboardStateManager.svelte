<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import { useMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors";
  import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateDataLoader";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import DashboardLoading from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardLoading.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy } from "svelte";

  export let exploreName: string;
  export let storageNamespacePrefix: string | undefined = undefined;
  export let bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<MetricsExplorerEntity> | null>
    | undefined = undefined;

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";
  $: exploreStore = useExploreState(exploreName);

  let dataLoader: DashboardStateDataLoader | undefined;
  $: if (metricsViewName) {
    dataLoader = new DashboardStateDataLoader(
      instanceId,
      metricsViewName,
      exploreName,
      storageNamespacePrefix,
      bookmarkOrTokenExploreState,
    );
  }

  let fullTimeRangeQuery:
    | ReturnType<typeof useMetricsViewTimeRange>
    | undefined;
  let stateSync: DashboardStateSync | undefined;
  $: if (dataLoader) {
    stateSync?.teardown();
    stateSync = new DashboardStateSync(
      instanceId,
      metricsViewName,
      exploreName,
      storageNamespacePrefix,
      dataLoader,
    );
    ({ fullTimeRangeQuery } = dataLoader);
  }

  let initExploreState:
    | CompoundQueryResult<MetricsExplorerEntity | undefined>
    | undefined;
  $: if (dataLoader) ({ initExploreState } = dataLoader);

  let error: HTTPError | null;
  let isLoading: boolean;
  $: if (initExploreState) {
    ({ isLoading, error } = $initExploreState as {
      isLoading: boolean;
      error: HTTPError | null;
    });
  }

  afterNavigate(({ from, to, type }) => {
    if (!from?.url || !to?.url || !stateSync) return;

    void stateSync.handleURLChange(to.url.searchParams, type);
  });

  onDestroy(() => {
    stateSync?.teardown();
  });

  // The timeRangeSummary is null when there are 0 rows of data
  // Notably, this happens when a security policy fully restricts a user from reading any data
  $: timeRangeSummaryIsNull =
    $fullTimeRangeQuery?.data?.timeRangeSummary?.min === null &&
    $fullTimeRangeQuery?.data?.timeRangeSummary?.max === null;
</script>

{#if isLoading}
  <DashboardLoading {isLoading} />
{:else if error}
  <ErrorPage
    statusCode={error.response?.status}
    header="Failed to load dashboard"
    detail={error.response?.data?.message ?? error.message}
  />
{:else if timeRangeSummaryIsNull}
  <ErrorPage
    header="Failed to load dashboard"
    body="This dashboard currently has no data to display. This may be due to access permissions."
  />
{:else if $exploreStore}
  <slot />
{/if}
