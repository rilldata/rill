<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useReport } from "../selectors";
  import MetadataLabel from "./MetadataLabel.svelte";
  import MetadataValue from "./MetadataValue.svelte";
  import { exportFormatToPrettyString } from "./utils";

  export let report: string;

  $: reportQuery = useReport($runtime.instanceId, report);

  $: metricViewsName =
    ($reportQuery.data &&
      JSON.parse($reportQuery.data.resource.report.spec.queryArgsJson)
        ?.metrics_view_name) ??
    "";
  $: dashboard = useDashboard($runtime.instanceId, metricViewsName);
  $: dashboardTitle = $dashboard.data?.metricsView.spec.title;

  $: exportLimit = $reportQuery.data.resource.report.spec.exportLimit;
</script>

{#if $reportQuery.data}
  <div class="flex flex-col gap-y-8 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <div class="flex gap-x-2 items-center">
        <h1 class="text-gray-800 text-base font-medium leading-none">
          {$reportQuery.data.resource.meta.name.name}
        </h1>
        <Tag>Report</Tag>
        <div class="grow" />
        <!-- TODO -->
        <!-- <Button type="primary" on:click={handleExportNow}>Export now</Button> -->
      </div>
    </div>

    <!-- Three columns of metadata -->
    <div class="flex flex-wrap gap-x-16 gap-y-3">
      <!-- Column 1 -->
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 grid-cols-[auto,1fr]">
        <!-- Dashboard -->
        <MetadataLabel>Dashboard</MetadataLabel>
        <MetadataValue>
          {dashboardTitle}
        </MetadataValue>
        <!-- Creator -->
        <MetadataLabel>Creator</MetadataLabel>
        <MetadataValue>
          {$reportQuery.data.resource.report.spec.annotations[
            "admin_owner_user_id"
          ] ?? "A project admin"}
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
            {exportLimit === "0" ? "No limit" : exportLimit}
          </MetadataValue>
        </div>
      </div>

      <!-- Column 3 -->
      <div class="flex flex-col gap-y-3">
        <!-- Frequency -->
        <div class="flex gap-x-6">
          <MetadataLabel>Frequency</MetadataLabel>
          <MetadataValue>
            {$reportQuery.data.resource.report.spec.refreshSchedule.cron}
          </MetadataValue>
        </div>
        <!-- Next run -->
        <div class="flex gap-x-6">
          <MetadataLabel>Next run</MetadataLabel>
          <MetadataValue>
            {new Date(
              $reportQuery.data.resource.report.state.nextRunOn
            ).toLocaleString("en-US", {
              month: "short",
              day: "numeric",
              year: "numeric",
              hour: "numeric",
              minute: "numeric",
              hour12: true,
            })}
          </MetadataValue>
        </div>
      </div>
    </div>
  </div>
{/if}
