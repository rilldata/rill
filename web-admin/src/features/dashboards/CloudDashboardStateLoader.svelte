<script lang="ts">
  import {
    createAdminServiceGetCurrentUser,
    type V1Project,
  } from "@rilldata/web-admin/client";
  import { getHomeBookmarkExploreState } from "@rilldata/web-admin/features/bookmarks/selectors";
  import { DashboardStateDataLoader } from "@rilldata/web-common/features/dashboards/state-managers/DashboardStateDataLoader";
  import { DashboardStateSync } from "@rilldata/web-common/features/dashboards/state-managers/DashboardStateSync";
  import DashboardLoading from "@rilldata/web-common/features/dashboards/url-state/DashboardLoading.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { QueriesStatus } from "@rilldata/web-common/runtime-client/QueriesStatus";
  import { onDestroy } from "svelte";
  import { derived } from "svelte/store";

  export let organization: string;
  export let project: V1Project;
  export let exploreName: string;

  $: ({ instanceId } = $runtime);
  $: exploreSpecQuery = useExploreValidSpec(instanceId, exploreName);
  $: metricsViewSpec = $exploreSpecQuery.data?.metricsView ?? {};
  $: exploreSpec = $exploreSpecQuery.data?.explore ?? {};
  $: metricsViewName = exploreSpec?.metricsView ?? "";
  $: exploreStore = useExploreState(exploreName);
  $: prefix = `${organization}__${project.name}__`;

  const userQuery = createAdminServiceGetCurrentUser();
  $: enableBookmarkQuery = Boolean($userQuery.data?.user);
  $: exploreStateFromHomeBookmarkQuery = getHomeBookmarkExploreState(
    project.id,
    instanceId,
    metricsViewName,
    exploreName,
    enableBookmarkQuery,
  );

  $: dataLoader = new DashboardStateDataLoader(
    instanceId,
    metricsViewName,
    exploreName,
    prefix,
    [derived(exploreStateFromHomeBookmarkQuery, (bq) => bq.data)],
  );

  let stateSync: DashboardStateSync | undefined;
  $: {
    stateSync?.teardown();
    stateSync = new DashboardStateSync(
      instanceId,
      metricsViewName,
      exploreName,
      prefix,
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
                query: dataLoader.fullTimeRangeQuery as any,
              },
            ]
          : []),
        ...(enableBookmarkQuery
          ? [
              {
                errorHeader: "Failed to fetch bookmarks for the explore.",
                query: exploreStateFromHomeBookmarkQuery as any,
              },
            ]
          : []),
      ],
      1000,
      5000,
    );
  }
  $: ({ loading, loadingForShortTime, loadingForLongTime, errors } =
    queriesStatus);
  $: firstError = $errors[0];

  onDestroy(() => {
    stateSync?.teardown();
    queriesStatus?.teardown();
  });

  $: console.log($exploreStore);
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
