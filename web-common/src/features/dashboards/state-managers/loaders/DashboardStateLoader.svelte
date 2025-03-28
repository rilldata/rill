<script lang="ts">
  import { afterNavigate } from "$app/navigation";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import type { SupportedCompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateDataLoader";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateSync";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
  import DashboardLoading from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardLoading.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { QueriesStatus } from "@rilldata/web-common/runtime-client/QueriesStatus";
  import { onDestroy } from "svelte";
  import { derived } from "svelte/store";

  export let exploreName: string;
  export let otherSourcesOfState: {
    errorHeader: string;
    query: SupportedCompoundQueryResult<
      Partial<MetricsExplorerEntity> | undefined,
      HTTPError
    >;
  }[] = [];

  const DASHBOARD_SHOW_SPINNER_THRESHOLD = 1000;
  const DASHBOARD_SHOW_LONG_LOAD_MESSAGE_THRESHOLD = 5000;

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $exploreSpecQuery.data?.metricsView ?? {};
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";
  $: exploreStore = useExploreState(exploreName);

  $: dataLoader = new DashboardStateDataLoader(
    instanceId,
    metricsViewName,
    exploreName,
    undefined,
    otherSourcesOfState.map(({ query }) => derived(query, (q) => q.data)),
  );

  let stateSync: DashboardStateSync | undefined;
  $: {
    stateSync?.teardown();
    stateSync = new DashboardStateSync(
      instanceId,
      metricsViewName,
      exploreName,
      undefined,
      dataLoader,
    );
  }

  let queriesStatus: QueriesStatus | undefined;
  $: if (dataLoader) {
    queriesStatus?.teardown();
    queriesStatus = new QueriesStatus(
      [
        {
          errorHeader: "Failed to explore.",
          query: dataLoader.validSpecQuery,
        },
        ...(metricsViewSpec.timeDimension
          ? [
              {
                errorHeader: "Failed to fetch time range for the explore.",
                query: dataLoader.fullTimeRangeQuery,
              },
            ]
          : []),
        ...otherSourcesOfState,
      ],
      DASHBOARD_SHOW_SPINNER_THRESHOLD,
      DASHBOARD_SHOW_LONG_LOAD_MESSAGE_THRESHOLD,
    );
  }
  $: ({ loading, loadingForShortTime, loadingForLongTime, errors } =
    queriesStatus!);
  $: firstError = $errors[0];

  afterNavigate(({ from, to, type }) => {
    if (!from?.url || !to?.url || !stateSync) return;

    void stateSync.handleURLChange(to.url.searchParams, type);
  });

  onDestroy(() => {
    stateSync?.teardown();
    queriesStatus?.teardown();
  });
</script>

{#if $loading}
  <DashboardLoading
    loadingForShortTime={$loadingForShortTime}
    loadingForLongTime={$loadingForLongTime}
  />
{:else if firstError}
  <ErrorPage
    statusCode={firstError.statusCode}
    header={firstError.header}
    body={firstError.body}
    detail={firstError.detail}
  />
{:else if $exploreStore}
  <slot />
{/if}
