<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import cronstrue from "cronstrue";
  import {
    createAdminServiceDeleteReport,
    createAdminServiceListProjectMembers,
  } from "../../../client";
  import { useReport, useReportDashboardName } from "../selectors";
  import MetadataLabel from "./MetadataLabel.svelte";
  import MetadataValue from "./MetadataValue.svelte";
  import { exportFormatToPrettyString, formatNextRunOn } from "./utils";

  export let organization: string;
  export let project: string;
  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);

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
      }
    );

  // Get owner's name
  const membersQuery = createAdminServiceListProjectMembers(
    organization,
    project
  );
  $: owner =
    $reportQuery.data &&
    $membersQuery.data &&
    $membersQuery.data.members.find(
      (member) =>
        member.userId ===
        $reportQuery.data.resource.report.spec.annotations[
          "admin_owner_user_id"
        ]
    );

  // Actions
  const queryClient = useQueryClient();
  const deleteReport = createAdminServiceDeleteReport();
  async function handleDeleteReport() {
    await $deleteReport.mutateAsync({
      organization,
      project,
      name: $reportQuery.data.resource.meta.name.name,
    });
    queryClient.invalidateQueries(
      getRuntimeServiceListResourcesQueryKey($runtime.instanceId)
    );
    goto(`/${organization}/${project}/-/reports`);
  }
</script>

{#if $reportQuery.data}
  <div class="flex flex-col gap-y-8 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-800 text-base font-medium leading-none">
          {$reportQuery.data.resource.report.spec.title}
        </h1>
        <div class="grow" />
        <!-- TODO: add more buttons & put delete report into a "..." icon button menu -->
        <!-- TODO: add an "are you sure?" confirmation dialog -->
        <Button
          type="secondary"
          on:click={handleDeleteReport}
          disabled={$deleteReport.isLoading}
        >
          Delete report
        </Button>
      </div>
    </div>

    <!-- Three columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-3">
      <!-- Column 1 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- Dashboard -->
        <MetadataLabel>Dashboard</MetadataLabel>
        <MetadataValue>
          <a href={`/${organization}/${project}/${$dashboardName.data}`}
            >{dashboardTitle}</a
          >
        </MetadataValue>
        <!-- Creator -->
        <MetadataLabel>Creator</MetadataLabel>
        <MetadataValue>
          {owner?.userName || "a project admin"}
        </MetadataValue>
        <!-- Recipients -->
        <MetadataLabel>Recipients</MetadataLabel>
        <MetadataValue>
          {$reportQuery.data.resource.report.spec.emailRecipients.length}
        </MetadataValue>
      </div>

      <!-- Column 2 -->
      <div class="flex flex-col gap-y-3">
        <!-- Format -->
        <div class="flex gap-x-6">
          <MetadataLabel>Format</MetadataLabel>
          <MetadataValue>
            {exportFormatToPrettyString(
              $reportQuery.data.resource.report.spec.exportFormat
            )}
          </MetadataValue>
        </div>
        <!-- Limit -->
        <div class="flex gap-x-6">
          <MetadataLabel>Limit</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data?.resource.report.spec.exportLimit === "0"
              ? "None"
              : $reportQuery.data?.resource.report.spec.exportLimit}
          </MetadataValue>
        </div>
      </div>

      <!-- Column 3 -->
      <div class="flex flex-col gap-y-3">
        <!-- Frequency -->
        <div class="flex gap-x-6">
          <MetadataLabel>Frequency</MetadataLabel>
          <MetadataValue>
            {humanReadableFrequency}
          </MetadataValue>
        </div>
        <!-- Next run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Next run</MetadataLabel>
          <MetadataValue>
            {formatNextRunOn(
              $reportQuery.data.resource.report.state.nextRunOn,
              $reportQuery.data.resource.report.spec.refreshSchedule.timeZone
            )}
          </MetadataValue>
        </div>
      </div>
    </div>
  </div>
{/if}
