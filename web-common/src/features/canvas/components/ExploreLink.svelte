<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/canvas/stores/types";
  import { useGetExploresForMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { mapQueryToDashboard } from "@rilldata/web-common/features/query-mappers/mapQueryToDashboard";
  import { getExplorePageUrlSearchParams } from "@rilldata/web-common/features/query-mappers/utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let metricsViewName: string;
  export let timeAndFilterStore: TimeAndFilterStore;

  const executionTime = new Date().toISOString();
  const queryName = "MetricsViewAggregation";

  let dashboardStateForComponent: ReturnType<typeof mapQueryToDashboard>;
  let isNavigating = false;
  let navigationError: string | null = null;

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: exploresQuery = useGetExploresForMetricsView(instanceId, metricsViewName);
  $: exploreDashboards = $exploresQuery.data ?? [];
  $: exploreName = exploreDashboards[0]?.meta?.name?.name ?? "";

  $: queryArgsJson = JSON.stringify({
    instanceId,
    metricsView: metricsViewName,
    timeRange: timeAndFilterStore.timeRange,
    comparisonTimeRange: timeAndFilterStore.comparisonTimeRange,
    where: timeAndFilterStore.where,
  });

  $: dashboardStateForComponent = mapQueryToDashboard(
    exploreName,
    queryName,
    queryArgsJson,
    executionTime,
    {},
  );

  async function gotoExplorePage() {
    navigationError = null;
    if (!$dashboardStateForComponent?.data) return;
    isNavigating = true;
    let url: URL;
    try {
      if (organization && project) {
        url = new URL(
          `/${organization}/${project}/explore/${encodeURIComponent(exploreName)}`,
          window.location.origin,
        );
      } else {
        url = new URL(
          `/explore/${encodeURIComponent(exploreName)}`,
          window.location.origin,
        );
      }
      url.search = (
        await getExplorePageUrlSearchParams(
          $dashboardStateForComponent?.data?.exploreName,
          $dashboardStateForComponent?.data?.exploreState,
        )
      ).toString();
      await goto(url.toString());
    } catch (err) {
      navigationError =
        err?.message || "Failed to navigate to Explore Dashboard.";
    } finally {
      isNavigating = false;
    }
  }
</script>

{#if exploreDashboards?.length > 0}
  <a
    on:click|preventDefault={() => gotoExplorePage()}
    class="inline-flex items-center gap-2"
    type="text"
  >
    {#if isNavigating}
      <Spinner status={EntityStatus.Running} size="1em" />
    {/if}
    Go to Explore Dashboard
  </a>
  {#if navigationError}
    <div class="flex flex-col gap-y-2 text-red-600">
      <h2 class="text-lg font-semibold">Unable to open Explore Dashboard</h2>
      <span>{navigationError}</span>
    </div>
  {/if}
  {#if $dashboardStateForComponent?.error && $dashboardStateForComponent?.isLoading}
    <div class="flex flex-col gap-y-2">
      <h2 class="text-lg font-semibold">Unable to open Explore Dashboard</h2>
      <span>{$dashboardStateForComponent.error}</span>
    </div>
  {/if}
{/if}
