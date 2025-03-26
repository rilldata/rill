<script lang="ts">
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/DashboardStateDataLoader";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/DashboardStateSync";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import DashboardLoading from "@rilldata/web-common/features/dashboards/url-state/DashboardLoading.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { QueriesStatus } from "@rilldata/web-common/runtime-client/QueriesStatus";
  import { onDestroy } from "svelte";

  export let exploreName: string;

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
    [],
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
      ],
      1000,
      5000,
    );
  }
  $: ({ loading, loadingForShortTime, loadingForLongTime, errors } =
    queriesStatus!);
  $: firstError = $errors[0];

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
