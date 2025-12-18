<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceDeleteAlert } from "@rilldata/web-admin/client";
  import EditAlert from "@rilldata/web-admin/features/alerts/EditAlert.svelte";
  import AlertFilterCriteria from "@rilldata/web-admin/features/alerts/metadata/AlertFilterCriteria.svelte";
  import AlertFilters from "@rilldata/web-admin/features/alerts/metadata/AlertFilters.svelte";
  import AlertOwnerBlock from "@rilldata/web-admin/features/alerts/metadata/AlertOwnerBlock.svelte";
  import { humaniseAlertSnoozeOption } from "@rilldata/web-admin/features/alerts/metadata/utils";
  import {
    useAlert,
    useAlertDashboardName,
    useAlertDashboardState,
    useIsAlertCreatedByCode,
  } from "@rilldata/web-admin/features/alerts/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataList from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataList.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { extractNotifier } from "@rilldata/web-admin/features/scheduled-reports/metadata/notifiers-utils";
  import { formatRefreshSchedule } from "@rilldata/web-admin/features/scheduled-reports/metadata/utils.ts";
  import { IconButton } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { hasValidMetricsViewTimeRange } from "@rilldata/web-common/features/dashboards/selectors.ts";
  import { getMappedExploreUrl } from "@rilldata/web-common/features/explore-mappers/get-mapped-explore-url.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import {
    getRuntimeServiceListResourcesQueryKey,
    type V1MetricsViewAggregationRequest,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let organization: string;
  export let project: string;
  export let alert: string;

  $: ({ instanceId } = $runtime);

  $: alertQuery = useAlert(instanceId, alert);
  $: isAlertCreatedByCode = useIsAlertCreatedByCode(instanceId, alert);

  // Get dashboard
  $: exploreName = useAlertDashboardName(instanceId, alert);
  $: validSpecResp = useExploreValidSpec(instanceId, $exploreName.data);
  $: exploreSpec = $validSpecResp.data?.explore;
  $: metricsViewName = exploreSpec?.metricsView;
  $: dashboardTitle = exploreSpec?.displayName || $exploreName.data;
  $: dashboardDoesNotExist = $validSpecResp.error?.response?.status === 404;

  $: exploreIsValid = hasValidMetricsViewTimeRange(
    instanceId,
    $exploreName.data,
  );

  $: alertSpec = $alertQuery.data?.resource?.alert?.spec;

  // Get human-readable frequency
  $: humanReadableFrequency = alertSpec?.refreshSchedule?.cron
    ? formatRefreshSchedule(alertSpec.refreshSchedule.cron)
    : "Whenever your data refreshes";

  $: queryArgsJson =
    (alertSpec?.resolverProperties?.query_args_json as string) ||
    alertSpec?.queryArgsJson ||
    "{}";
  $: queryName =
    alertSpec?.queryName ||
    (alertSpec?.resolverProperties?.query_name as string);
  $: metricsViewAggregationRequest = JSON.parse(
    queryArgsJson,
  ) as V1MetricsViewAggregationRequest;

  $: dashboardState = useAlertDashboardState(instanceId, alertSpec);

  $: snoozeLabel = humaniseAlertSnoozeOption(alertSpec);

  $: emailNotifier = extractNotifier(alertSpec?.notifiers, "email");
  $: slackNotifier = extractNotifier(alertSpec?.notifiers, "slack");

  $: exploreUrl = getMappedExploreUrl(
    {
      exploreName: $exploreName.data,
      queryName,
      queryArgsJson,
    },
    {
      exploreProtoState: alertSpec?.annotations?.web_open_state,
    },
    {
      instanceId,
      organization,
      project,
    },
  );

  // Actions
  const queryClient = useQueryClient();
  const deleteAlert = createAdminServiceDeleteAlert();

  async function handleDeleteAlert() {
    await $deleteAlert.mutateAsync({
      org: organization,
      project,
      name: $alertQuery.data.resource.meta.name.name,
    });
    await queryClient.invalidateQueries({
      queryKey: getRuntimeServiceListResourcesQueryKey(instanceId),
    });
    // goto only after invalidate is complete
    goto(`/${organization}/${project}/-/alerts`);
  }
</script>

{#if alertSpec}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-gray-500 font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            {#if $alertQuery.data}
              <AlertOwnerBlock
                {organization}
                {project}
                ownerId={alertSpec.annotations["admin_owner_user_id"]}
              />
            {/if}
          </svelte:fragment>
        </ProjectAccessControls>
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-700 text-lg font-bold" aria-label="Alert name">
          {alertSpec.displayName}
        </h1>
        <div class="grow" />
        {#if !$isAlertCreatedByCode.data}
          <EditAlert {alertSpec} disabled={!$exploreIsValid} />
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton ariaLabel="Alert context menu">
                <ThreeDot size="16px" />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item on:click={handleDeleteAlert}>
                Delete Alert
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
      </div>
    </div>

    <!-- Five columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Dashboard -->
      <div class="flex flex-col gap-y-3" aria-label="Alert dashboard name">
        {#if dashboardTitle}
          <MetadataLabel>Dashboard</MetadataLabel>
          <MetadataValue>
            {#if dashboardDoesNotExist}
              <div class="flex items-center gap-x-1">
                {dashboardTitle}
                <Tooltip distance={8}>
                  <CancelCircle size="16px" className="text-red-500" />
                  <TooltipContent slot="tooltip-content">
                    Dashboard does not exist
                  </TooltipContent>
                </Tooltip>
              </div>
            {:else}
              <a href={$exploreUrl}>
                {dashboardTitle}
              </a>
            {/if}
          </MetadataValue>
        {:else}
          <MetadataLabel>Name</MetadataLabel>
          <MetadataValue>
            {$alertQuery.data?.resource?.meta?.name?.name}
          </MetadataValue>
        {/if}
      </div>

      <!-- Split by dimension -->
      <div class="flex flex-col gap-y-3" aria-label="Alert split by dimension">
        <MetadataLabel>Split by dimension</MetadataLabel>
        <MetadataValue>
          {metricsViewAggregationRequest?.dimensions?.[0]?.name ?? "None"}
        </MetadataValue>
      </div>

      <!-- Schedule: TODO: change based on non UI settings -->
      <div class="flex flex-col gap-y-3" aria-label="Alert schedule">
        <MetadataLabel>Schedule</MetadataLabel>
        <MetadataValue>{humanReadableFrequency}</MetadataValue>
      </div>

      <!-- Snooze -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Snooze</MetadataLabel>
        <MetadataValue>{snoozeLabel}</MetadataValue>
      </div>
    </div>

    <!-- Filters -->
    <AlertFilters
      {metricsViewName}
      filters={metricsViewAggregationRequest?.where}
      dimensionsWithInlistFilter={$dashboardState.data
        ?.dimensionsWithInlistFilter ?? []}
      timeRange={metricsViewAggregationRequest?.timeRange}
      comparisonTimeRange={metricsViewAggregationRequest?.comparisonTimeRange}
    />

    <!-- Criteria -->
    <AlertFilterCriteria
      filters={metricsViewAggregationRequest?.having}
      comparisonTimeRange={metricsViewAggregationRequest?.comparisonTimeRange}
    />

    <!-- Slack notification -->
    {#if slackNotifier}
      <MetadataList
        data={[...slackNotifier.channels, ...slackNotifier.users]}
        label="Slack notifications"
      />
    {/if}

    <!-- Email notifications -->
    {#if emailNotifier}
      <MetadataList
        data={emailNotifier.recipients}
        label="Email notifications"
      />
    {/if}
  </div>
{/if}
