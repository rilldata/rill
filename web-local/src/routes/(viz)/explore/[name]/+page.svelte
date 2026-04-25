<script lang="ts">
  import { onNavigate } from "$app/navigation";
  import {
    DashboardBannerID,
    DashboardBannerPriority,
  } from "@rilldata/web-common/components/banner/constants";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import DashboardBuilding from "@rilldata/web-common/features/dashboards/DashboardBuilding.svelte";
  import { resetSelectedMockUserAfterNavigate } from "@rilldata/web-common/features/dashboards/granular-access-policies/resetSelectedMockUserAfterNavigate";
  import { selectedMockUserStore } from "@rilldata/web-common/features/dashboards/granular-access-policies/stores";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { useProjectParser } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import {
    useExploreWithPolling,
    isExploreReconcilingForFirstTime,
    isExploreErrored,
  } from "@rilldata/web-common/features/explores/selectors";
  import VisualExploreEditing from "@rilldata/web-common/features/workspaces/VisualExploreEditing.svelte";
  import {
    extractErrorStatusCode,
    isNotFoundError,
  } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  const runtimeClient = useRuntimeClient();

  export let data: PageData;
  $: ({ exploreName } = data);

  resetSelectedMockUserAfterNavigate(queryClient, runtimeClient);

  $: exploreResource = useExploreWithPolling(runtimeClient, exploreName);

  $: validSpec = $exploreResource.data?.explore?.explore?.state?.validSpec;
  $: metricsViewName = $exploreResource.data?.metricsView?.meta?.name
    ?.name as string;
  $: measures = validSpec?.measures ?? [];

  $: filePaths = [
    ...($exploreResource.data?.explore?.meta?.filePaths ?? []),
    ...($exploreResource.data?.metricsView?.meta?.filePaths ?? []),
  ];

  // Inline-explore editing: when this dashboard is synthesized from a
  // metrics view's `explore:` block (or v0 default), the inspector edits
  // that block in-place inside the metrics_view YAML.
  $: explore = $exploreResource.data?.explore?.explore;
  $: inlineExplore = !!validSpec?.definedInMetricsView;
  $: metricsFilePath = $exploreResource.data?.metricsView?.meta?.filePaths?.[0];
  $: showInlineInspector =
    !$previewModeStore && inlineExplore && !!metricsFilePath;
  $: inlineFileArtifact = metricsFilePath
    ? fileArtifacts.getFileArtifact(metricsFilePath)
    : undefined;
  $: inlineAutoSave = inlineFileArtifact ? inlineFileArtifact.autoSave : null;

  $: projectParserQuery = useProjectParser(queryClient, runtimeClient, {
    enabled: $selectedMockUserStore?.admin,
  });

  $: hasBanner = !!validSpec?.banner;

  $: if (hasBanner) {
    eventBus.emit("add-banner", {
      id: DashboardBannerID,
      priority: DashboardBannerPriority,
      message: {
        type: "default",
        message: validSpec?.banner ?? "",
        iconType: "alert",
      },
    });
  }

  $: dashboardFileHasParseError =
    $projectParserQuery.data?.projectParser?.state?.parseErrors?.filter(
      (error) => filePaths.includes(error.filePath as string),
    );

  $: isDashboardNotFound =
    !$exploreResource.data &&
    $exploreResource.isError &&
    isNotFoundError($exploreResource.error);

  $: mockUserHasNoAccess =
    $selectedMockUserStore && isNotFoundError($exploreResource.error);

  $: homeHref = $previewModeStore ? "/dashboards" : "/";

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

{#if $exploreResource.isPending && !$exploreResource.data}
  <DashboardBuilding />
{:else if mockUserHasNoAccess}
  <ErrorPage
    statusCode={extractErrorStatusCode($exploreResource.error)}
    header="This user can't access this dashboard"
    body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
    href={homeHref}
  />
{:else if isDashboardNotFound}
  <ErrorPage statusCode={404} header="Dashboard not found" href={homeHref} />
{:else if $exploreResource.isSuccess}
  {#if isExploreReconcilingForFirstTime($exploreResource.data)}
    <DashboardBuilding />
  {:else if isExploreErrored($exploreResource.data)}
    <ErrorPage
      header="Error building dashboard"
      body={$exploreResource.data?.explore?.meta?.reconcileError ??
        "An unknown error occurred while building the dashboard."}
      href={homeHref}
    />
  {:else if dashboardFileHasParseError && dashboardFileHasParseError.length > 0}
    <ErrorPage
      header="Error parsing dashboard"
      body="Please check your dashboard's YAML file for errors."
      href={homeHref}
    />
  {:else if measures.length === 0 && $selectedMockUserStore !== null}
    <ErrorPage
      statusCode={extractErrorStatusCode($exploreResource.error)}
      header="Error fetching dashboard"
      body="No measures available"
      href={homeHref}
    />
  {:else if metricsViewName}
    <div class="h-full flex overflow-hidden">
      {#key exploreName}
        <StateManagersProvider {metricsViewName} {exploreName} visualEditing>
          <div class="flex-1 overflow-hidden">
            <DashboardStateManager {exploreName}>
              <Dashboard {metricsViewName} {exploreName} />
            </DashboardStateManager>
          </div>
          {#if showInlineInspector && inlineFileArtifact && $inlineAutoSave !== null}
            <VisualExploreEditing
              fileArtifact={inlineFileArtifact}
              {exploreName}
              exploreResource={explore}
              {metricsViewName}
              viewingDashboard
              autoSave={$inlineAutoSave}
              switchView={() => {}}
              propertyPathPrefix={["explore"]}
            />
          {/if}
        </StateManagersProvider>
      {/key}
    </div>
  {/if}
{/if}
