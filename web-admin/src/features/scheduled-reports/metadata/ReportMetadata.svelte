<script lang="ts">
  import { goto } from "$app/navigation";
  import MetadataList from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataList.svelte";
  import { extractNotifier } from "@rilldata/web-admin/features/scheduled-reports/metadata/notifiers-utils";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import EditScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/EditScheduledReportDialog.svelte";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import cronstrue from "cronstrue";
  import { createAdminServiceDeleteReport } from "../../../client";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import {
    useIsReportCreatedByCode,
    useReport,
    useReportDashboardName,
  } from "../selectors";
  import MetadataLabel from "./MetadataLabel.svelte";
  import MetadataValue from "./MetadataValue.svelte";
  import ReportOwnerBlock from "./ReportOwnerBlock.svelte";
  import RunNowButton from "./RunNowButton.svelte";
  import { exportFormatToPrettyString, formatNextRunOn } from "./utils";

  export let organization: string;
  export let project: string;
  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);
  $: isReportCreatedByCode = useIsReportCreatedByCode(
    $runtime.instanceId,
    report,
  );

  // Get dashboard
  $: dashboardName = useReportDashboardName($runtime.instanceId, report);
  $: dashboard = useDashboard($runtime.instanceId, $dashboardName.data);
  $: dashboardTitle = $dashboard.data?.metricsView.spec.title;

  // Get human-readable frequency
  $: humanReadableFrequency =
    $reportQuery.data &&
    cronstrue.toString(
      $reportQuery.data.resource.report.spec.refreshSchedule.cron,
      {
        verbose: true,
      },
    );

  $: emailNotifier =
    $reportQuery.data &&
    extractNotifier($reportQuery.data.resource.report.spec.notifiers, "email");

  // Actions
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();

  let showEditReportDialog = false;
  function handleEditReport() {
    showEditReportDialog = true;
  }

  async function handleDeleteReport() {
    await $deleteReport.mutateAsync({
      organization,
      project,
      name: $reportQuery.data.resource.meta.name.name,
    });
    queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId),
    );
    goto(`/${organization}/${project}/-/reports`);
  }
</script>

{#if $reportQuery.data}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-gray-500 font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            {#if $reportQuery.data}
              <ReportOwnerBlock
                {organization}
                {project}
                ownerId={$reportQuery.data.resource.report.spec.annotations[
                  "admin_owner_user_id"
                ]}
              />
            {/if}
          </svelte:fragment>
        </ProjectAccessControls>
        <!-- Format -->
        <span>
          {exportFormatToPrettyString(
            $reportQuery.data.resource.report.spec.exportFormat,
          )} •
        </span>
        <!-- Limit -->
        <span>
          {$reportQuery.data?.resource.report.spec.exportLimit === "0"
            ? "No row limit"
            : `${$reportQuery.data?.resource.report.spec.exportLimit} row limit`}
        </span>
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-700 text-lg font-bold">
          {$reportQuery.data.resource.report.spec.title}
        </h1>
        <div class="grow" />
        <RunNowButton {organization} {project} {report} />
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

    <!-- Three columns of metadata -->
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

      <!-- Frequency -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Repeats</MetadataLabel>
        <MetadataValue>
          {humanReadableFrequency}
        </MetadataValue>
      </div>

      <!-- Next run -->
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Next run</MetadataLabel>
        <MetadataValue>
          {formatNextRunOn(
            $reportQuery.data.resource.report.state.nextRunOn,
            $reportQuery.data.resource.report.spec.refreshSchedule.timeZone,
          )}
        </MetadataValue>
      </div>
    </div>

    <!-- Recipients -->
    {#if emailNotifier}
      <MetadataList data={emailNotifier.recipients} label="Recipients" />
    {/if}
  </div>
{/if}

{#if $reportQuery.data}
  <EditScheduledReportDialog
    open={showEditReportDialog}
    reportSpec={$reportQuery.data.resource.report.spec}
    on:close={() => (showEditReportDialog = false)}
  />
{/if}
