<script lang="ts">
  import { goto } from "$app/navigation";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import EditScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/EditScheduledReportDialog.svelte";
  import {
    V1Report,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import cronstrue from "cronstrue";
  import {
    V1ExportFormat,
    createAdminServiceDeleteReport,
  } from "../../../client";
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
    report
  );

  // Get dashboard
  $: dashboardName = useReportDashboardName($runtime.instanceId, report);
  $: dashboard = useDashboard($runtime.instanceId, $dashboardName.data);
  $: dashboardTitle = $dashboard.data?.metricsView?.spec?.title;

  $: cron = $reportQuery.data?.resource?.report?.spec?.refreshSchedule?.cron;
  // Get human-readable frequency
  $: humanReadableFrequency =
    cron &&
    cronstrue.toString(cron, {
      verbose: true,
    });

  // Actions
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();

  let showEditReportDialog = false;
  function handleEditReport() {
    showEditReportDialog = true;
  }

  async function handleDeleteReport() {
    const name = $reportQuery?.data?.resource?.meta?.name?.name;
    if (!name) {
      // if name is not ready, then we can't delete the report
      return;
    }
    await $deleteReport.mutateAsync({
      organization,
      project,
      name,
    });
    queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId)
    );
    goto(`/${organization}/${project}/-/reports`);
  }

  /**
   * Casts V1Report data. This assumes that if a V1Report is present,
   * then all the fields are present. This cast fn can be removed once
   * API types correctly reflect which fields are optional.
   * @param data
   */
  function reportCast(report: V1Report) {
    return {
      admin_owner_user_id: report.spec?.annotations?.[
        "admin_owner_user_id"
      ] as string,
      exportFormat: report?.spec?.exportFormat as V1ExportFormat,
      exportLimit: report?.spec?.exportLimit as string,
      title: report?.spec?.title as string,
      emailRecipients: report?.spec?.emailRecipients as string[],
      nextRunOn: report?.state?.nextRunOn as string,
      timeZone: report?.spec?.refreshSchedule?.timeZone as string,
    };
  }

  $: v1report = $reportQuery.data?.resource?.report
    ? reportCast($reportQuery.data.resource.report)
    : undefined;
</script>

{#if v1report}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-gray-500 font-semibold">
        <!-- Author -->
        <ProjectAccessControls {organization} {project}>
          <svelte:fragment slot="manage-project">
            <ReportOwnerBlock
              {organization}
              {project}
              ownerId={v1report.admin_owner_user_id}
            />
          </svelte:fragment>
        </ProjectAccessControls>
        <!-- Format -->
        <span>
          {exportFormatToPrettyString(v1report.exportFormat)} •
        </span>
        <!-- Limit -->
        <span>
          {v1report.exportLimit === "0"
            ? "No row limit"
            : `${v1report.exportLimit} row limit`}
        </span>
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-700 text-lg font-bold">
          {v1report.title}
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
          {formatNextRunOn(v1report.nextRunOn, v1report.timeZone)}
        </MetadataValue>
      </div>
    </div>

    <!-- Recipients -->
    <div class="flex flex-col gap-y-3">
      <MetadataLabel
        >Recipients ({v1report.emailRecipients.length})</MetadataLabel
      >
      <div class="flex flex-wrap gap-2">
        {#each v1report.emailRecipients as recipient}
          <Tag>
            {recipient}
          </Tag>
        {/each}
      </div>
    </div>
  </div>
{/if}

{#if $reportQuery?.data?.resource?.report?.spec}
  <EditScheduledReportDialog
    open={showEditReportDialog}
    reportSpec={$reportQuery?.data?.resource?.report?.spec}
    on:close={() => (showEditReportDialog = false)}
  />
{/if}
