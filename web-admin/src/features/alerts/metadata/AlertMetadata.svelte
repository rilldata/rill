<script lang="ts">
  import AlertFilterCriteria from "@rilldata/web-admin/features/alerts/metadata/AlertFilterCriteria.svelte";
  import {
    useAlert,
    useAlertDashboardName,
    useIsAlertCreatedByCode,
  } from "@rilldata/web-admin/features/alerts/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import EmailRecipients from "@rilldata/web-admin/features/scheduled-reports/metadata/EmailRecipients.svelte";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import ReportOwnerBlock from "@rilldata/web-admin/features/scheduled-reports/metadata/ReportOwnerBlock.svelte";
  import { IconButton } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import type { V1MetricsViewAggregationRequest } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import AlertFilters from "@rilldata/web-admin/features/alerts/metadata/AlertFilters.svelte";

  export let organization: string;
  export let project: string;
  export let alert: string;

  $: alertQuery = useAlert($runtime.instanceId, alert);
  $: isReportCreatedByCode = useIsAlertCreatedByCode(
    $runtime.instanceId,
    alert,
  );

  // Get dashboard
  $: dashboardName = useAlertDashboardName($runtime.instanceId, alert);
  $: dashboard = useDashboard($runtime.instanceId, $dashboardName.data);
  $: dashboardTitle = $dashboard.data?.metricsView.spec.title;

  $: metricsViewAggregationRequest = JSON.parse(
    $alertQuery.data.resource.alert.spec.queryArgsJson,
  ) as V1MetricsViewAggregationRequest;

  // TODO: delete and edit
  function handleEditReport() {}

  function handleDeleteReport() {}
</script>

{#if $alertQuery.data}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-gray-500 font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            {#if $alertQuery.data}
              <ReportOwnerBlock
                {organization}
                {project}
                ownerId={$alertQuery.data.resource.alert.spec.annotations[
                  "admin_owner_user_id"
                ]}
              />
            {/if}
          </svelte:fragment>
        </ProjectAccessControls>
      </div>
      <!-- Header row 2 -->
      <div class="text-xl text-gray-500 font-bold">
        {$alertQuery.data.resource.alert.spec.title}
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-700 text-lg font-bold">
          {$alertQuery.data.resource.report.spec.title}
        </h1>
        <div class="grow" />
        {#if !$isReportCreatedByCode.data}
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton>
                <ThreeDot size="16px" />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="start">
              <DropdownMenu.Item on:click={handleEditReport}>
                Edit report
              </DropdownMenu.Item>
              <DropdownMenu.Item on:click={handleDeleteReport}>
                Delete report
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        {/if}
      </div>
    </div>

    <!-- Five columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Dashboard -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Dashboard</MetadataLabel>
        <MetadataValue>
          <a href={`/${organization}/${project}/${$dashboardName.data}`}
            >{dashboardTitle}</a
          >
        </MetadataValue>
      </div>

      <!-- Split by dimension -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Split by dimension</MetadataLabel>
        <MetadataValue>
          {metricsViewAggregationRequest?.dimensions[0]?.name}
        </MetadataValue>
      </div>

      <!-- Split by time grain -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Split by time grain</MetadataLabel>
        <MetadataValue>TODO</MetadataValue>
      </div>

      <!-- Schedule -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Schedule</MetadataLabel>
        <MetadataValue>TODO</MetadataValue>
      </div>

      <!-- Snooze -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Snooze</MetadataLabel>
        <MetadataValue>TODO</MetadataValue>
      </div>
    </div>

    <!-- Filters -->
    <AlertFilters
      metricsViewName={$dashboardName.data}
      filters={metricsViewAggregationRequest?.where}
    />

    <!-- Criteria -->
    <AlertFilterCriteria
      metricsViewName={$dashboardName.data}
      filters={metricsViewAggregationRequest?.having}
    />

    <!-- Recipients -->
    <EmailRecipients
      emailRecipients={$alertQuery.data.resource.alert.spec.emailRecipients}
    />
  </div>
{/if}
