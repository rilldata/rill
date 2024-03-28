<script lang="ts">
  import { goto } from "$app/navigation";
  import { createAdminServiceDeleteAlert } from "@rilldata/web-admin/client";
  import AlertFilterCriteria from "@rilldata/web-admin/features/alerts/metadata/AlertFilterCriteria.svelte";
  import AlertFilters from "@rilldata/web-admin/features/alerts/metadata/AlertFilters.svelte";
  import AlertOwnerBlock from "@rilldata/web-admin/features/alerts/metadata/AlertOwnerBlock.svelte";
  import { humaniseAlertSnoozeOption } from "@rilldata/web-admin/features/alerts/metadata/utils";
  import {
    useAlert,
    useAlertDashboardName,
    useIsAlertCreatedByCode,
  } from "@rilldata/web-admin/features/alerts/selectors";
  import ProjectAccessControls from "@rilldata/web-admin/features/projects/ProjectAccessControls.svelte";
  import EmailRecipients from "@rilldata/web-admin/features/scheduled-reports/metadata/EmailRecipients.svelte";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { IconButton } from "@rilldata/web-common/components/button";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import EditAlertDialog from "@rilldata/web-common/features/alerts/EditAlertDialog.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    getRuntimeServiceListResourcesQueryKey,
    type V1MetricsViewAggregationRequest,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let organization: string;
  export let project: string;
  export let alert: string;

  $: alertQuery = useAlert($runtime.instanceId, alert);
  $: isAlertCreatedByCode = useIsAlertCreatedByCode($runtime.instanceId, alert);

  // Get dashboard
  $: dashboardName = useAlertDashboardName($runtime.instanceId, alert);
  $: dashboard = useDashboard($runtime.instanceId, $dashboardName.data);
  $: dashboardTitle =
    $dashboard.data?.metricsView.spec.title || $dashboardName.data;

  $: metricsViewAggregationRequest = JSON.parse(
    $alertQuery.data?.resource?.alert?.spec?.queryArgsJson ?? "{}",
  ) as V1MetricsViewAggregationRequest;

  $: snoozeLabel = humaniseAlertSnoozeOption(
    $alertQuery.data?.resource?.alert?.spec,
  );

  // Actions
  const queryClient = useQueryClient();
  const deleteAlert = createAdminServiceDeleteAlert();

  let showEditAlertDialog = false;
  function handleEditAlert() {
    showEditAlertDialog = true;
  }

  async function handleDeleteAlert() {
    await $deleteAlert.mutateAsync({
      organization,
      project,
      name: $alertQuery.data.resource.meta.name.name,
    });
    await queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
    );
    // goto only after invalidate is complete
    goto(`/${organization}/${project}/-/alerts`);
  }
</script>

{#if $alertQuery.data?.resource?.alert?.spec}
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
                ownerId={$alertQuery.data.resource.alert.spec.annotations[
                  "admin_owner_user_id"
                ]}
              />
            {/if}
          </svelte:fragment>
        </ProjectAccessControls>
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-700 text-lg font-bold">
          {$alertQuery.data.resource.alert.spec.title}
        </h1>
        <div class="grow" />
        {#if !$isAlertCreatedByCode.data}
          <Button type="secondary" on:click={handleEditAlert}>Edit</Button>
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton>
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

      <!-- Schedule: TODO: change based on non UI settings -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Schedule</MetadataLabel>
        <MetadataValue>Whenever your data refreshes</MetadataValue>
      </div>

      <!-- Snooze -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Snooze</MetadataLabel>
        <MetadataValue>{snoozeLabel}</MetadataValue>
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
      emailRecipients={$alertQuery.data.resource.alert.spec.notifySpec.notifiers.find(
        (n) => n.email,
      )?.email?.recipients ?? []}
    />
  </div>
{/if}

{#if $alertQuery.data && $dashboard.data?.metricsView.spec}
  <EditAlertDialog
    open={showEditAlertDialog}
    alertSpec={$alertQuery.data.resource.alert.spec}
    on:close={() => (showEditAlertDialog = false)}
    metricsViewName={$dashboardName.data}
  />
{/if}
