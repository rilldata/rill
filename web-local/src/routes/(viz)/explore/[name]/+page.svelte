<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import ExploreChat from "@rilldata/web-common/features/chat/ExploreChat.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;
  $: ({ metricsView, explore, exploreName } = data);

  resetSelectedMockUserAfterNavigate(queryClient);

  $: metricsViewName = metricsView?.meta?.name?.name as string;

  $: ({ instanceId } = $runtime);

  $: filePaths = [
    ...(explore.meta?.filePaths ?? []),
    ...(metricsView.meta?.filePaths ?? []),
  ];
  $: exploreQuery = useExploreValidSpec(instanceId, exploreName);
  $: measures = $exploreQuery.data?.explore?.measures ?? [];
  $: projectParserQuery = useProjectParser(queryClient, instanceId, {
    enabled: $selectedMockUserStore?.admin,
  });

  $: hasBanner = !!$exploreQuery.data?.explore?.banner;

  $: if (hasBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: $exploreQuery.data?.explore?.banner ?? "",
        iconType: "alert",
      },
    });
  }

  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );
  $: mockUserHasNoAccess =
    $selectedMockUserStore && $exploreQuery.error?.response?.status === 404;

  onNavigate(({ from, to }) => {
    const changedDashboard =
      !from || !to || from?.params?.name !== to?.params?.name;
    // Clear out any dashboard banners
    if (hasBanner && changedDashboard) {
      eventBus.emit("remove-banner", DashboardBannerID);
    }
  });
</script>

<svelte:head>
  <title>Rill Developer | {exploreName}</title>
</svelte:head>

{#if measures.length === 0 && $selectedMockUserStore !== null}
  <ErrorPage
    statusCode={$exploreQuery.error?.response?.status}
    header="Error fetching dashboard"
    body="No measures available"
  />
  <!-- Handle errors from dashboard YAML edits from an external IDE -->
{:else if dashboardFileHasParseError && dashboardFileHasParseError.length > 0}
  <ErrorPage
    header="Error parsing dashboard"
    body="Please check your dashboard's YAML file for errors."
  />
{:else if mockUserHasNoAccess}
  <ErrorPage
    statusCode={$exploreQuery.error?.response?.status}
    header="This user can't access this dashboard"
    body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
  />
{:else}
  {#key exploreName}
    <div class="flex h-full overflow-hidden">
      <div class="flex-1 overflow-hidden">
        <StateManagersProvider {metricsViewName} {exploreName}>
          <DashboardStateManager {exploreName}>
            <Dashboard {metricsViewName} {exploreName} />
          </DashboardStateManager>
        </StateManagersProvider>
      </div>
      <ExploreChat />
    </div>
  {/key}
{/if}
